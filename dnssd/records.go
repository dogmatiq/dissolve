package dnssd

import (
	"net"
	"time"

	"github.com/miekg/dns"
)

// DefaultTTL is the default TTL to use for DNS records.
const DefaultTTL = 2 * time.Minute

// NewPTRRecord returns the PTR record for a service instance.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-4.1
func NewPTRRecord(i Instance) *dns.PTR {
	return &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   InstanceEnumerationDomain(i.ServiceType, i.Domain) + ".",
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(i.TTL),
		},
		Ptr: i.FullyQualifiedName() + ".",
	}
}

// NewSRVRecord returns the SRV record for a service instance.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-5.
func NewSRVRecord(i Instance) *dns.SRV {
	return &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   i.FullyQualifiedName() + ".",
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

// NewTXTRecord returns the TXT record for a service instance.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-6.
func NewTXTRecord(i Instance) *dns.TXT {
	return &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   i.FullyQualifiedName() + ".",
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(i.TTL),
		},
		Txt: i.Attributes.ToTXT(),
	}
}

// NewARecord returns an A record for a service instance.
//
// ip must be an IPv4 address, or an IPv4 address expresses as an IPv6 address.
func NewARecord(i Instance, ip net.IP) *dns.A {
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
func NewAAAARecord(i Instance, ip net.IP) *dns.AAAA {
	ip = ip.To16()
	if ip == nil {
		panic("IP address is not a valid IP address")
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
