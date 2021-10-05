package dnssd

import (
	"context"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// DefaultUnicastQueryTimeout is the default time to allow for unicast DNS
// queries to be served.
const DefaultUnicastQueryTimeout = 500 * time.Millisecond

// UnicastServer is a conventional (unicast) DNS server designed specifically
// for serving DNS-SD records.
//
// It does not support recursive DNS queries (i.e, it can not forward requests
// for unknown domains to upstream DNS servers).
type UnicastServer struct {
	// Timeout is the amount of time to allow for each DNS query.
	//
	// If it is non-positive, DefaultUnicastQueryTimeout is used instead.
	Timeout time.Duration

	m sync.RWMutex

	// services store information about the records related to a specific
	// service type.
	//
	// The key is the instance enumeration domain for a specific service type
	// and domain.
	services map[string]*serviceRecords

	// instances stores information about the records related to specific
	// service instances.
	//
	// The key is the fully-qualified service name.
	instances map[string]*instanceRecords

	// records is a map of domain to the records within that domain. The inner
	// map maps record type to the records of that type.
	records map[string]map[uint16][]dns.RR
}

type serviceRecords struct {
	typeEnumRecord *dns.PTR
	instanceCount  int
}

type instanceRecords struct {
	serviceRecords *serviceRecords
	records        []dns.RR
}

// Advertise starts advertising a DNS-SD service instance.
//
// addresses is a set of optional IP addresses. If provided, the server will
// also respond with the relevant A and AAAA requests when it receives a query
// for the hostname in i.TargetHost.
//
// Typically, these records would be served by a separate domain name server
// that is authoratative for the internet domain name used in i.TargetHost.
func (s *UnicastServer) Advertise(i ServiceInstance, options ...AdvertiseOption) {
	name := ServiceInstanceName(i.Instance, i.ServiceType, i.Domain)
	records := NewRecords(i, options...)

	s.m.Lock()
	defer s.m.Unlock()

	if s.instances == nil {
		s.services = map[string]*serviceRecords{}
		s.instances = map[string]*instanceRecords{}
		s.records = map[string]map[uint16][]dns.RR{}
	} else {
		s.removeInstance(name)
	}

	enumDomain := InstanceEnumerationDomain(i.ServiceType, i.Domain)

	sr, ok := s.services[enumDomain]
	if ok {
		sr.instanceCount++
	} else {
		sr = &serviceRecords{
			NewServiceTypePTRRecord(i.ServiceType, i.Domain, 0),
			1,
		}

		s.services[enumDomain] = sr
		s.addRecord(sr.typeEnumRecord)
	}

	s.instances[name] = &instanceRecords{sr, records}

	for _, rr := range records {
		s.addRecord(rr)
	}
}

// Remove stops advertising a DNS-SD service instance.
func (s *UnicastServer) Remove(i ServiceInstance) {
	name := ServiceInstanceName(i.Instance, i.ServiceType, i.Domain)

	s.m.Lock()
	defer s.m.Unlock()

	s.removeInstance(name)
}

func (s *UnicastServer) removeInstance(name string) {
	ir, ok := s.instances[name]
	if !ok {
		return
	}

	ir.serviceRecords.instanceCount--

	if ir.serviceRecords.instanceCount == 0 {
		s.removeRecord(ir.serviceRecords.typeEnumRecord)
		delete(s.services, ir.serviceRecords.typeEnumRecord.Ptr)
	}

	for _, rr := range ir.records {
		s.removeRecord(rr)
	}

	delete(s.instances, name)
}

// addRecord adds a record to the DNS server. It assumes s.m is already locked
// for writing.
func (s *UnicastServer) addRecord(rr dns.RR) {
	h := rr.Header()

	domainRecords := s.records[h.Name]
	if domainRecords == nil {
		domainRecords = map[uint16][]dns.RR{}
		s.records[h.Name] = domainRecords
	}

	domainRecords[h.Rrtype] = append(domainRecords[h.Rrtype], rr)
}

// removeRecord removes a record from the DNS server. It assumes s.m is already
// locked for writing.
func (s *UnicastServer) removeRecord(rr dns.RR) {
	h := rr.Header()

	domainRecords := s.records[h.Name]
	typeRecords := domainRecords[h.Rrtype]

	for i, x := range typeRecords {
		if x != rr {
			continue
		}

		lastIndex := len(typeRecords) - 1

		if lastIndex == 0 {
			// If this is the last remaining record of this type we can remove
			// the entire slice from typeRecords.
			delete(domainRecords, h.Rrtype)

			// Likewise, if the domain contains no more records of any kind,
			// remove the entire domainRecords map from s.records.
			if len(domainRecords) == 0 {
				delete(s.records, h.Name)
			}

			return
		}

		// Otherwise, we want to remove the i'th element. We do this efficiently
		// by moving the last element to position i, then shrinking the slice by
		// one.
		typeRecords[i] = typeRecords[lastIndex]
		typeRecords[lastIndex] = nil // prevent memory leak from reference in underlying array
		typeRecords = typeRecords[:lastIndex]

		domainRecords[h.Rrtype] = typeRecords

		return
	}
}

// Run runs the server until ctx is canceled or an error occurs.
func (s *UnicastServer) Run(ctx context.Context, network, address string) error {
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = DefaultUnicastQueryTimeout
	}

	server := &dns.Server{
		Net:          network,
		Addr:         address,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		Handler: dns.HandlerFunc(
			func(w dns.ResponseWriter, req *dns.Msg) {
				defer w.Close()

				if res, ok := s.buildResponse(req); ok {
					_ = w.WriteMsg(res)
				}
			},
		),
	}

	// Create a context we can cancel when we exit so we can always signal
	// server.Shutdown() to be called.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a channel that is used to signal when server.Shutdown() has
	// completed.
	done := make(chan struct{})

	go func() {
		defer close(done)     // signal shutdown goroutine has ended
		<-ctx.Done()          // wait for cancellation
		_ = server.Shutdown() // shutdown server
	}()

	// Always wait for the shutdown goroutine to finish before actually
	// returning.
	defer func() { <-done }()

	err := server.ListenAndServe()

	// If the context was canceled we don't care about whatever listener-related
	// error is reported to us, just tell the caller about the context error.
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return err
}

// buildResponse builds the response to send in reply to the given request.
func (s *UnicastServer) buildResponse(req *dns.Msg) (*dns.Msg, bool) {
	// We only support queries with exactly one question. The RFC allows for
	// multiple, but in practice this is nonsensical.
	//
	// See https://stackoverflow.com/questions/4082081/requesting-a-and-aaaa-records-in-single-dns-query/4085631#4085631
	// See https://datatracker.ietf.org/doc/html/rfc1035
	if len(req.Question) != 1 {
		return nil, false
	}

	q := req.Question[0]

	res := &dns.Msg{}
	res.SetReply(req)
	res.Authoritative = true
	res.RecursionAvailable = false

	s.m.RLock()
	defer s.m.RUnlock()

	// Copy the records to res.Answer. We don't want to reference the slices
	// inside s.records as they may be modified as soon as s.m is unlocked.
	if q.Qtype == dns.TypeANY {
		for _, records := range s.records[q.Name] {
			res.Answer = append(res.Answer, records...)
		}
	} else {
		res.Answer = append(res.Answer, s.records[q.Name][q.Qtype]...)
	}

	if len(res.Answer) == 0 && len(s.records[q.Name]) == 0 {
		res.Rcode = dns.RcodeNameError
	}

	return res, true
}
