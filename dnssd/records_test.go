package dnssd_test

import (
	"net"

	. "github.com/dogmatiq/dissolve/dnssd"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Context("DNS records", func() {
	var instance Instance

	BeforeEach(func() {
		instance = Instance{
			Name:        "Living Room TV.",
			ServiceType: "_airplay._tcp",
			Domain:      "local",
			TargetHost:  "host.example.org",
			TargetPort:  12345,
			Priority:    10,
			Weight:      20,
		}

		instance.Attributes.Set("<key>", []byte("<value>"))
	})

	Describe("func NewRecords()", func() {
		It("returns all of the records required to announce a service instance", func() {
			records := NewRecords(instance)

			Expect(records).To(ConsistOf(
				&dns.PTR{
					Hdr: dns.RR_Header{
						Name:   `_airplay._tcp.local.`,
						Rrtype: dns.TypePTR,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Ptr: `Living Room TV\.._airplay._tcp.local.`,
				},
				&dns.SRV{
					Hdr: dns.RR_Header{
						Name:   `Living Room TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeSRV,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Target:   "host.example.org.",
					Port:     12345,
					Priority: 10,
					Weight:   20,
				},
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   `Living Room TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Txt: []string{"<key>=<value>"},
				},
			))
		})

		It("adds A and AAAA records if IP addresses are passed", func() {
			records := NewRecords(
				instance,
				net.IPv4(192, 168, 20, 1),
				net.ParseIP("fe80::1ce5:3c8b:36f:53cf"),
			)

			Expect(records).To(ContainElements(
				&dns.A{
					Hdr: dns.RR_Header{
						Name:   `host.example.org.`,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					A: net.IPv4(192, 168, 20, 1).To4(),
				},
				&dns.AAAA{
					Hdr: dns.RR_Header{
						Name:   `host.example.org.`,
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					AAAA: net.IPv4(192, 168, 20, 1).To16(),
				},
				&dns.AAAA{
					Hdr: dns.RR_Header{
						Name:   `host.example.org.`,
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					AAAA: net.IP{0xfe, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1c, 0xe5, 0x3c, 0x8b, 0x03, 0x6f, 0x53, 0xcf},
				},
			))
		})
	})

	Describe("func NewPTRRecord()", func() {
		It("returns the expected PTR record", func() {
			rec := NewPTRRecord(instance)

			Expect(rec).To(Equal(
				&dns.PTR{
					Hdr: dns.RR_Header{
						Name:   `_airplay._tcp.local.`,
						Rrtype: dns.TypePTR,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Ptr: `Living Room TV\.._airplay._tcp.local.`,
				},
			))
		})
	})

	Describe("func NewSRVRecord()", func() {
		It("returns the expected SRV record", func() {
			rec := NewSRVRecord(instance)

			Expect(rec).To(Equal(
				&dns.SRV{
					Hdr: dns.RR_Header{
						Name:   `Living Room TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeSRV,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Target:   "host.example.org.",
					Port:     12345,
					Priority: 10,
					Weight:   20,
				},
			))
		})
	})

	Describe("func NewTXTRecord()", func() {
		It("returns the expected TXT record", func() {
			rec := NewTXTRecord(instance)

			Expect(rec).To(Equal(
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   `Living Room TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Txt: []string{"<key>=<value>"},
				},
			))
		})
	})

	Describe("func NewARecord()", func() {
		It("returns the expected A record for an IPv4 address", func() {
			rec := NewARecord(instance, net.IPv4(192, 168, 20, 1).To4())

			Expect(rec).To(Equal(
				&dns.A{
					Hdr: dns.RR_Header{
						Name:   `host.example.org.`,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					A: net.IPv4(192, 168, 20, 1).To4(),
				},
			))
		})

		It("returns the expected A record for an IPv4 address encoded within an IPv6 address.", func() {
			rec := NewARecord(instance, net.IPv4(192, 168, 20, 1).To16())

			Expect(rec).To(Equal(
				&dns.A{
					Hdr: dns.RR_Header{
						Name:   `host.example.org.`,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					A: net.IPv4(192, 168, 20, 1).To4(),
				},
			))
		})
	})

	Describe("func NewAAAARecord()", func() {
		It("returns the expected AAAA record for an IPv6 address", func() {
			rec := NewAAAARecord(instance, net.ParseIP("fe80::1ce5:3c8b:36f:53cf"))

			Expect(rec).To(Equal(
				&dns.AAAA{
					Hdr: dns.RR_Header{
						Name:   `host.example.org.`,
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					AAAA: net.IP{0xfe, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1c, 0xe5, 0x3c, 0x8b, 0x03, 0x6f, 0x53, 0xcf},
				},
			))
		})

		It("returns the expected AAAA record for an IPv4 address", func() {
			rec := NewAAAARecord(instance, net.IPv4(192, 168, 20, 1))

			Expect(rec).To(Equal(
				&dns.AAAA{
					Hdr: dns.RR_Header{
						Name:   `host.example.org.`,
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					AAAA: net.IPv4(192, 168, 20, 1).To16(),
				},
			))
		})
	})

	Describe("func NewServiceTypePTRRecord()", func() {
		It("returns the expected PTR record", func() {
			rec := NewServiceTypePTRRecord("_airplay._tcp", "local", 0)

			Expect(rec).To(Equal(
				&dns.PTR{
					Hdr: dns.RR_Header{
						Name:   `_services._dns-sd._udp.local.`,
						Rrtype: dns.TypePTR,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Ptr: `_airplay._tcp.local.`,
				},
			))
		})
	})
})
