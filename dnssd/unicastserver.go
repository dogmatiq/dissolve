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
//
// It implements the [Advertiser] interface.
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

var _ Advertiser = (*UnicastServer)(nil)

type serviceRecords struct {
	typeEnumRecord *dns.PTR
	instanceCount  int
}

type instanceRecords struct {
	serviceRecords *serviceRecords
	records        []dns.RR
}

// Advertise starts advertising a DNS-SD service instance.
func (s *UnicastServer) Advertise(
	_ context.Context,
	inst ServiceInstance,
	options ...AdvertiseOption,
) (bool, error) {
	name := AbsoluteServiceInstanceName(inst.Name, inst.ServiceType, inst.Domain)
	records := NewRecords(inst, options...)

	s.m.Lock()
	defer s.m.Unlock()

	if s.hasRecords(records) {
		return false, nil
	}

	if s.instances == nil {
		s.services = map[string]*serviceRecords{}
		s.instances = map[string]*instanceRecords{}
		s.records = map[string]map[uint16][]dns.RR{}
	} else {
		s.removeInstance(name)
	}

	enumDomain := AbsoluteInstanceEnumerationDomain(inst.ServiceType, inst.Domain)

	sr, ok := s.services[enumDomain]
	if ok {
		sr.instanceCount++
	} else {
		sr = &serviceRecords{
			NewServiceTypePTRRecord(inst.ServiceType, inst.Domain, 0),
			1,
		}

		s.services[enumDomain] = sr
		s.addRecord(sr.typeEnumRecord)
	}

	s.instances[name] = &instanceRecords{sr, records}

	for _, rr := range records {
		s.addRecord(rr)
	}

	return true, nil
}

// Unadvertise stops advertising a DNS-SD service instance.
func (s *UnicastServer) Unadvertise(
	_ context.Context,
	inst ServiceInstance,
) (bool, error) {
	name := AbsoluteServiceInstanceName(inst.Name, inst.ServiceType, inst.Domain)

	s.m.Lock()
	defer s.m.Unlock()

	return s.removeInstance(name), nil
}

func (s *UnicastServer) removeInstance(name string) bool {
	ir, ok := s.instances[name]
	if !ok {
		return false
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

	return true
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

func (s *UnicastServer) hasRecords(records []dns.RR) bool {
	for _, rr := range records {
		if !s.hasRecord(rr) {
			return false
		}
	}

	return true
}

func (s *UnicastServer) hasRecord(rr dns.RR) bool {
	h := rr.Header()

	for _, x := range s.records[h.Name][h.Rrtype] {
		if x.String() == rr.String() {
			return true
		}
	}

	return false
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
	// See https://www.rfc-editor.org/rfc/rfc1035
	if len(req.Question) != 1 {
		return nil, false
	}

	q := req.Question[0]

	res := &dns.Msg{}
	res.SetReply(req)
	res.Authoritative = true
	res.RecursionAvailable = false

	if q.Qclass != dns.ClassINET && q.Qclass != dns.ClassANY {
		res.Rcode = dns.RcodeNameError
		return res, true
	}

	s.m.RLock()
	defer s.m.RUnlock()

	records := s.records[q.Name]

	if len(records) == 0 {
		res.Rcode = dns.RcodeNameError
		return res, true
	}

	// Always use a copy of the records in res.Answer.
	//
	// We don't want to reference the original slice(s) from s.records as they
	// may be modified as soon as s.m is unlocked.
	if q.Qtype == dns.TypeANY {
		for _, recs := range records {
			res.Answer = append(res.Answer, recs...)
		}
	} else {
		res.Answer = append([]dns.RR{}, records[q.Qtype]...)
	}

	return res, true
}
