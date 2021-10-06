package dnssd_test

import (
	"net"

	. "github.com/dogmatiq/dissolve/dnssd"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Context("DNS records", func() {
	var instance ServiceInstance

	BeforeEach(func() {
		instance = ServiceInstance{
			Instance:    "Living Room TV.",
			ServiceType: "_airplay._tcp",
			Domain:      "local",
			TargetHost:  "host.example.org",
			TargetPort:  12345,
			Priority:    10,
			Weight:      20,
			Attributes:  []Attributes{{}},
		}

		instance.Attributes[0].Set("<key>", []byte("<value>"))
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
					Ptr: `Living\ Room\ TV\.._airplay._tcp.local.`,
				},
				&dns.SRV{
					Hdr: dns.RR_Header{
						Name:   `Living\ Room\ TV\.._airplay._tcp.local.`,
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
						Name:   `Living\ Room\ TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Txt: []string{"<key>=<value>"},
				},
			))
		})

		It("adds A and AAAA records if the WithIPAddress() option is used", func() {
			records := NewRecords(
				instance,
				WithIPAddress(net.IPv4(192, 168, 20, 1)),
				WithIPAddress(net.ParseIP("fe80::1ce5:3c8b:36f:53cf")),
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
					Ptr: `Living\ Room\ TV\.._airplay._tcp.local.`,
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
						Name:   `Living\ Room\ TV\.._airplay._tcp.local.`,
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

	Describe("func NewTXTRecords()", func() {
		It("returns the expected TXT records", func() {
			rec := NewTXTRecords(instance)

			Expect(rec).To(ConsistOf(
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   `Living\ Room\ TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Txt: []string{"<key>=<value>"},
				},
			))
		})

		It("returns the expected TXT records when there are multiple attribute collections", func() {
			var attrs Attributes
			attrs.Set("<key-2>", []byte("<value-2>"))
			instance.Attributes = append(instance.Attributes, attrs)

			rec := NewTXTRecords(instance)

			Expect(rec).To(ConsistOf(
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   `Living\ Room\ TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Txt: []string{"<key>=<value>"},
				},
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   `Living\ Room\ TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Txt: []string{"<key-2>=<value-2>"},
				},
			))
		})

		It("ignores empty attribute collections", func() {
			instance.Attributes = append(instance.Attributes, Attributes{})
			rec := NewTXTRecords(instance)

			Expect(rec).To(ConsistOf(
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   `Living\ Room\ TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
					Txt: []string{"<key>=<value>"},
				},
			))
		})

		It("returns a single empty record if there are no attributes", func() {
			instance.Attributes = nil
			rec := NewTXTRecords(instance)

			Expect(rec).To(ConsistOf(
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   `Living\ Room\ TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
				},
			))
		})

		It("returns a single empty record if there are only empty attribute collections", func() {
			instance.Attributes = []Attributes{{}, {}}
			rec := NewTXTRecords(instance)

			Expect(rec).To(ConsistOf(
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   `Living\ Room\ TV\.._airplay._tcp.local.`,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    120,
					},
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

		It("panics if given an IPv4 address", func() {
			Expect(func() {
				NewAAAARecord(instance, net.IPv4(192, 168, 20, 1).To4())
			}).To(PanicWith("can not produce an AAAA record for an IPv4 address"))
		})

		It("panics if given an IPv4 address encoded within an IPv6 address", func() {
			Expect(func() {
				NewAAAARecord(instance, net.IPv4(192, 168, 20, 1).To16())
			}).To(PanicWith("can not produce an AAAA record for an IPv4 address"))
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
