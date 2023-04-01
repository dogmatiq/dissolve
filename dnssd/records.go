package dnssd

import (
	"net"
	"time"

	"github.com/miekg/dns"
)

// DefaultTTL is the default TTL to use for DNS records.
const DefaultTTL = 2 * time.Minute

// NewRecords returns the set of DNS-SD records used to announce the given
// service instance.
func NewRecords(i ServiceInstance, options ...AdvertiseOption) []dns.RR {
	opts := resolveAdvertiseOptions(options)

	records := []dns.RR{
		NewPTRRecord(i),
		NewSRVRecord(i),
	}

	for _, rr := range NewTXTRecords(i) {
		records = append(records, rr)
	}

	for _, subType := range opts.ServiceSubTypes {
		records = append(records, NewServiceSubTypePTRRecord(i, subType))
	}

	for _, ip := range opts.IPAddresses {
		if ip.To4() != nil {
			records = append(records, NewARecord(i, ip))
		} else if ip.To16() != nil {
			records = append(records, NewAAAARecord(i, ip))
		}
	}

	return records
}

// NewPTRRecord returns the PTR record for a service instance.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-4.1
func NewPTRRecord(i ServiceInstance) *dns.PTR {
	return &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   InstanceEnumerationDomain(i.ServiceType, i.Domain) + ".",
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(i.TTL),
		},
		Ptr: AbsoluteServiceInstanceName(i.Name, i.ServiceType, i.Domain) + ".",
	}
}

// NewSRVRecord returns the SRV record for a service instance.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-5.
func NewSRVRecord(i ServiceInstance) *dns.SRV {
	return &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   AbsoluteServiceInstanceName(i.Name, i.ServiceType, i.Domain) + ".",
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(i.TTL),
		},
		Priority: i.Priority,
		Weight:   i.Weight,
		Target:   i.TargetHost + ".",
		Port:     i.TargetPort,
	}
}

// NewTXTRecords returns TXT records containing a service instance's attributes.
//
// It returns one TXT record for each non-empty set of attributes in
// i.Attributes.
//
// If there are no attributes, it returns a single empty TXT record.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-6.
// See https://www.rfc-editor.org/rfc/rfc6763#section-6.8.
func NewTXTRecords(i ServiceInstance) []*dns.TXT {
	header := dns.RR_Header{
		Name:   AbsoluteServiceInstanceName(i.Name, i.ServiceType, i.Domain) + ".",
		Rrtype: dns.TypeTXT,
		Class:  dns.ClassINET,
		Ttl:    ttlInSeconds(i.TTL),
	}

	var records []*dns.TXT

	for _, attrs := range i.Attributes {
		if !attrs.IsEmpty() {
			records = append(
				records,
				&dns.TXT{
					Hdr: header,
					Txt: attrs.ToTXT(),
				},
			)
		}
	}

	// Each instance must have at least one TXT record, so even if i.Attributes
	// is empty, or only contained empty collections, we still add an empty text
	// record.
	if len(records) == 0 {
		records = append(
			records,
			&dns.TXT{
				Hdr: header,
				Txt: []string{""},
			},
		)
	}

	return records
}

// NewServiceSubTypePTRRecord returns a PTR record used to advertise a service
// was providing a specific service sub-type.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-7.1.
func NewServiceSubTypePTRRecord(i ServiceInstance, subType string) *dns.PTR {
	return &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   SelectiveInstanceEnumerationDomain(subType, i.ServiceType, i.Domain) + ".",
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(i.TTL),
		},
		Ptr: AbsoluteServiceInstanceName(i.Name, i.ServiceType, i.Domain) + ".",
	}
}

// NewARecord returns an A record for a service instance.
//
// ip must be an IPv4 address, or an IPv4 address expresses as an IPv6 address.
func NewARecord(i ServiceInstance, ip net.IP) *dns.A {
	ip = ip.To4()
	if ip == nil {
		panic("IP address is not a valid IPv4 address")
	}

	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   i.TargetHost + ".",
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(i.TTL),
		},
		A: append(net.IP{}, ip...), // clone IP
	}
}

// NewAAAARecord returns an A record for a service instance.
//
// ip must be a valid IPv4 or IPv6 address.
func NewAAAARecord(i ServiceInstance, ip net.IP) *dns.AAAA {
	if ip.To4() != nil {
		panic("can not produce an AAAA record for an IPv4 address")
	}

	ip = ip.To16()
	if ip == nil {
		panic("IP address is not a valid IPv6 address")
	}

	return &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   i.TargetHost + ".",
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(i.TTL),
		},
		AAAA: append(net.IP{}, ip...), // clone IP
	}
}

// NewServiceTypePTRRecord returns the PTR record for a service type.
//
// These records are sent in response to a service type enumeration request.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-9
func NewServiceTypePTRRecord(serviceType, domain string, ttl time.Duration) *dns.PTR {
	return &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   TypeEnumerationDomain(domain) + ".",
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(ttl),
		},
		Ptr: InstanceEnumerationDomain(serviceType, domain) + ".",
	}
}

// ttlInSeconds returns TTL as the number of whole seconds for use within a DNS
// record.
//
// if ttl is non-positive, DefaultTTL is used instead.
func ttlInSeconds(ttl time.Duration) uint32 {
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	return uint32(ttl.Seconds())
}
