package dnssd

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
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
// See https://datatracker.ietf.org/doc/html/rfc6763#section-4.1.
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
// See https://datatracker.ietf.org/doc/html/rfc6763#section-4.1.
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
// See https://datatracker.ietf.org/doc/html/rfc6763#section-4.1.
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
		if res.Rcode == dns.RcodeNameError || res.Rcode == dns.RcodeSuccess {
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
