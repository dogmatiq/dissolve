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
		NewTXTRecord(i),
	}

	for _, subType := range opts.ServiceSubTypes {
		records = append(records, NewServiceSubTypePTRRecord(i, subType))
	}

	for _, ip := range opts.IPAddresses {
		if ip.To4() != nil {
			records = append(records, NewARecord(i, ip))
		}

		if ip.To16() != nil {
			records = append(records, NewAAAARecord(i, ip))
		}
	}

	return records
}

// NewPTRRecord returns the PTR record for a service instance.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-4.1
func NewPTRRecord(i ServiceInstance) *dns.PTR {
	return &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   InstanceEnumerationDomain(i.ServiceType, i.Domain) + ".",
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(i.TTL),
		},
		Ptr: ServiceInstanceName(i.Instance, i.ServiceType, i.Domain) + ".",
	}
}

// NewSRVRecord returns the SRV record for a service instance.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-5.
func NewSRVRecord(i ServiceInstance) *dns.SRV {
	return &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   ServiceInstanceName(i.Instance, i.ServiceType, i.Domain) + ".",
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

// NewTXTRecord returns a TXT record containing a service instance's attributes.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-6.
func NewTXTRecord(i ServiceInstance) *dns.TXT {
	return &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   ServiceInstanceName(i.Instance, i.ServiceType, i.Domain) + ".",
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(i.TTL),
		},
		Txt: i.Attributes.ToTXT(),
	}
}

// NewServiceSubTypePTRRecord returns a PTR record used to advertise a service
// was providing a specific service sub-type.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-7.1.
func NewServiceSubTypePTRRecord(i ServiceInstance, subType string) *dns.PTR {
	return &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   SelectiveInstanceEnumerationDomain(subType, i.ServiceType, i.Domain) + ".",
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    ttlInSeconds(i.TTL),
		},
		Ptr: ServiceInstanceName(i.Instance, i.ServiceType, i.Domain) + ".",
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

// NewServiceTypePTRRecord returns the PTR record for a service type.
//
// These records are sent in response to a service type enumeration request.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-9
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
