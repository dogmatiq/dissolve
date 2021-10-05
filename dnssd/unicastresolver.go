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

	client := r.Client
	if client == nil {
		client = &dns.Client{}
	}

	if d, ok := ctx.Deadline(); ok {
		if client == r.Client {
			// The user has provided a specific client to use. We take a copy
			// client so we can manipulate client.DialTimeout without causing a
			// data race.
			client = &dns.Client{
				Net:            r.Client.Net,
				UDPSize:        r.Client.UDPSize,
				TLSConfig:      r.Client.TLSConfig,
				Dialer:         r.Client.Dialer,
				Timeout:        r.Client.Timeout,
				ReadTimeout:    r.Client.ReadTimeout,
				WriteTimeout:   r.Client.WriteTimeout,
				TsigSecret:     r.Client.TsigSecret,
				TsigProvider:   r.Client.TsigProvider,
				SingleInflight: r.Client.SingleInflight,
			}
		}

		client.DialTimeout = time.Until(d)
	}

	req := &dns.Msg{}
	req.SetQuestion(name+".", questionType)

	for _, s := range r.Config.Servers {
		addr := net.JoinHostPort(s, r.Config.Port)
		res, _, err := client.Exchange(req, addr)

		// Manually check for context cancelation, as client.Exchange() does not
		// support it and client.ExchangeContext() is avoided due to a lack of
		// support for context cancelation (only deadlines) and the fact that it
		// clobbers the user's provided dialer, which aside from not behaving as
		// configured is potentially a data race.
		if ctx.Err() != nil {
			return nil, false, ctx.Err()
		}

		// We could not query this server, move on to the next one.
		if err != nil {
			continue
		}

		// Server was contactable but did not have a response for this query.
		if res == nil {
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
