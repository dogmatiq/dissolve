package dnssd

import (
	"context"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/sync/errgroup"
)

// UnicastResolver makes DNS-SD queries using unicast DNS requests.
//
// This is a relatively low-level interface that allows performing each type of
// DNS-SD query type separately.
type UnicastResolver struct {
	Client *dns.Client
	Config *dns.ClientConfig
}

// EnumerateServiceTypes finds all of the service types advertised within a
// single domain.
//
// It returns a slice containing the discovered service types, without the
// domain suffix.  This is the "<service>" portion of the "service instance
// name", For example "_http._tcp".
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-4.1.
func (r *UnicastResolver) EnumerateServiceTypes(
	ctx context.Context,
	domain string,
) ([]string, error) {
	res, ok, err := r.query(
		ctx,
		TypeEnumerationDomain(domain),
		dns.TypePTR,
	)
	if !ok || err != nil {
		return nil, err
	}

	suffix := "." + domain + "."
	serviceTypes := make([]string, 0, len(res.Answer))

	for _, rr := range res.Answer {
		if ptr, ok := rr.(*dns.PTR); ok {
			serviceType := strings.TrimSuffix(ptr.Ptr, suffix)
			if serviceType != ptr.Ptr {
				serviceTypes = append(serviceTypes, serviceType)
			}
		}
	}

	return serviceTypes, nil
}

// EnumerateInstances finds all of the instances of a given service type that
// are advertised within a single domain.
//
// This operation is also known as as "browsing".
//
// serviceType is the type of service to enumerate, for example "_http._tcp".
//
// It returns a slice of the instance names. This is the "<instance>" portion of
// the "service instance name", for example, "Boardroom Printer".
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-4.1.
func (r *UnicastResolver) EnumerateInstances(
	ctx context.Context,
	serviceType, domain string,
) ([]string, error) {
	res, ok, err := r.query(
		ctx,
		InstanceEnumerationDomain(serviceType, domain),
		dns.TypePTR,
	)
	if !ok || err != nil {
		return nil, err
	}

	instances := make([]string, 0, len(res.Answer))

	for _, rr := range res.Answer {
		if ptr, ok := rr.(*dns.PTR); ok {
			instance, _, err := ParseInstance(ptr.Ptr)
			if err == nil {
				instances = append(instances, instance)
			}
		}
	}

	return instances, nil
}

// EnumerateInstancesBySubType finds all of the instances of a given service
// sub-type that are advertised within a single domain.
//
// This operation is also known as "selective instance enumeration" or less
// commonly "selective browsing" or "sub-type browsing".
//
// subType is the specific service sub-type, such as "_printer". serviceType is
// the type of service to enumerate, for example "_http._tcp".
//
// It returns a slice of the instance names. This is the "<instance>" portion of
// the "service instance name", for example, "Boardroom Printer".
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-4.1.
func (r *UnicastResolver) EnumerateInstancesBySubType(
	ctx context.Context,
	subType, serviceType, domain string,
) ([]string, error) {
	res, ok, err := r.query(
		ctx,
		SelectiveInstanceEnumerationDomain(subType, serviceType, domain),
		dns.TypePTR,
	)
	if !ok || err != nil {
		return nil, err
	}

	instances := make([]string, 0, len(res.Answer))

	for _, rr := range res.Answer {
		if ptr, ok := rr.(*dns.PTR); ok {
			instance, _, err := ParseInstance(ptr.Ptr)
			if err == nil {
				instances = append(instances, instance)
			}
		}
	}

	return instances, nil
}

// LookupInstance looks up the details about a specific service instance.
//
// instance and serviceType are the "<instance>" and "<service>" portions of the
// instance name, for example "Boardroom Printer" and "_http._tcp", respectively.
//
// ok is false if the instance can not be respolved.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-4.1.
func (r *UnicastResolver) LookupInstance(
	ctx context.Context,
	instance, serviceType, domain string,
) (_ ServiceInstance, ok bool, _ error) {
	queryName := AbsoluteServiceInstanceName(instance, serviceType, domain)
	responses := make(chan *dns.Msg, 2)

	// Note that we make separate queries for SRV and TXT records. We do this
	// (rather than using an ANY query) as there is no requirement within the
	// DNS specification that servers respond with ALL records in response to an
	// ANY query.
	//
	// This common misconception is explained in the Multicast DNS RFC at
	// https://www.rfc-editor.org/rfc/rfc6762#section-6.5.
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		res, ok, err := r.query(ctx, queryName, dns.TypeSRV)
		if ok {
			responses <- res
		}
		return err
	})

	g.Go(func() error {
		res, ok, err := r.query(ctx, queryName, dns.TypeTXT)
		if ok {
			responses <- res
		}
		return err
	})

	if err := g.Wait(); err != nil {
		return ServiceInstance{}, false, err
	}

	close(responses)

	i := ServiceInstance{
		ServiceInstanceName: ServiceInstanceName{
			Name:        instance,
			ServiceType: serviceType,
			Domain:      domain,
		},
		TTL: math.MaxInt64,
	}

	var hasSRV, hasTXT bool

	for res := range responses {
		for _, rr := range res.Answer {
			ttl := time.Duration(rr.Header().Ttl) * time.Second
			if ttl < i.TTL {
				i.TTL = ttl
			}

			switch rr := rr.(type) {
			case *dns.SRV:
				hasSRV = true
				unpackSRV(&i, rr)
			case *dns.TXT:
				hasTXT = true
				if err := unpackTXT(&i, rr); err != nil {
					return ServiceInstance{}, false, err
				}
			}
		}
	}

	return i, hasSRV && hasTXT, nil
}

// unpackSRV unpacks information from a SRV record into i.
func unpackSRV(i *ServiceInstance, rr *dns.SRV) {
	i.TargetHost = strings.TrimSuffix(rr.Target, ".")
	i.TargetPort = rr.Port
	i.Priority = rr.Priority
	i.Weight = rr.Weight
}

// unpackSRV unpacks information from a TXT record into i.
func unpackTXT(i *ServiceInstance, rr *dns.TXT) error {
	var attrs Attributes

	for _, pair := range rr.Txt {
		var err error
		attrs, _, err = attrs.WithTXT(pair)
		if err != nil {
			return fmt.Errorf("unable to parse TXT record: %w", err)
		}
	}

	if !attrs.IsEmpty() {
		i.Attributes = append(i.Attributes, attrs)
	}

	return nil
}

// query performs a DNS query against all of the servers in r.Config.
func (r *UnicastResolver) query(
	ctx context.Context,
	name string,
	questionType uint16,
) (*dns.Msg, bool, error) {
	if r.Config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(r.Config.Timeout)*time.Second)
		defer cancel()
	}

	req := &dns.Msg{}
	req.SetQuestion(name+".", questionType)

	for _, s := range r.Config.Servers {
		if ctx.Err() != nil {
			return nil, false, ctx.Err()
		}

		addr := net.JoinHostPort(s, r.Config.Port)
		res, ok := r.queryServer(ctx, addr, req)

		// Server was not contactable or had no response for this query.
		if !ok {
			continue
		}

		// The server responded authoratatively, even if it was only to indicate
		// that this domain or record type does not exist.
		if res.Rcode == dns.RcodeNameError {
			break
		}

		// The server had an answer to this query.
		if res.Rcode == dns.RcodeSuccess {
			return res, true, nil
		}
	}

	// None of the servers had a result for this query.
	return nil, false, nil
}

// query performs a DNS query against all of the servers in r.Config.
func (r *UnicastResolver) queryServer(
	ctx context.Context,
	addr string,
	req *dns.Msg,
) (*dns.Msg, bool) {
	client := r.Client
	if client == nil {
		client = &dns.Client{}
	}

	conn, err := client.Dial(addr)
	if err != nil {
		return nil, false
	}

	// Create a context that is always canceled when we are finished with this
	// server.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Close the connection when the context is canceled, regardless of whether
	// it's before or after the exchange is complete. This also terminates the
	// exchange if the parent ctx is canceled for any reason, including reaching
	// its deadline.
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	res, _, _ := client.ExchangeWithConn(req, conn)
	return res, res != nil
}
