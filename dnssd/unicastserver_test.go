package dnssd_test

import (
	"context"
	"net"
	"time"

	. "github.com/dogmatiq/dissolve/dnssd"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("UnicastServer", func() {
	var (
		ctx                             context.Context
		cancel                          context.CancelFunc
		instanceA, instanceB, instanceC ServiceInstance
		server                          *UnicastServer
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)

		instanceA = ServiceInstance{
			ServiceInstanceName: ServiceInstanceName{
				Name:        "Instance A",
				ServiceType: "_http._tcp",
				Domain:      "example.org",
			},
			TargetHost: "a.example.com",
			TargetPort: 12345,
			Priority:   10,
			Weight:     20,
			Attributes: AttributeCollection{
				NewAttributes().
					WithPair("<key>", []byte("<instance-a>")),
			},
		}

		instanceB = ServiceInstance{
			ServiceInstanceName: ServiceInstanceName{
				Name:        "Instance B",
				ServiceType: "_http._tcp",
				Domain:      "example.org",
			},
			TargetHost: "b.example.com",
			TargetPort: 12345,
			Priority:   10,
			Weight:     20,
			Attributes: AttributeCollection{
				NewAttributes().
					WithPair("<key>", []byte("<instance-b0>")),
				NewAttributes().
					WithPair("<key>", []byte("<instance-b1>")),
			},
		}

		instanceC = ServiceInstance{
			ServiceInstanceName: ServiceInstanceName{
				Name:        "Instance C",
				ServiceType: "_other._udp",
				Domain:      "example.org",
			},
			TargetHost: "c.example.com",
			TargetPort: 12345,
			Priority:   10,
			Weight:     20,
		}

		server = &UnicastServer{}

		changed, err := server.Advertise(
			ctx,
			instanceA,
			WithServiceSubType("_printer"),
		)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(changed).To(BeTrue())

		changed, err = server.Advertise(
			ctx,
			instanceB,
			WithIPAddress(net.IPv4(192, 168, 20, 1)),
			WithIPAddress(net.ParseIP("fe80::1ce5:3c8b:36f:53cf")),
		)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(changed).To(BeTrue())

		changed, err = server.Advertise(
			ctx,
			instanceC,
		)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(changed).To(BeTrue())
	})

	AfterEach(func() {
		cancel()
	})

	Context("DNS responses", func() {
		var (
			client *dns.Client
			errors chan error
		)

		BeforeEach(func() {
			client = &dns.Client{}
			errors = make(chan error, 1)

			go func() {
				errors <- server.Run(ctx, "udp", "127.0.0.1:65353")
			}()

			// Fudge-factor to allow the server time to start.
			time.Sleep(100 * time.Millisecond)
		})

		AfterEach(func() {
			cancel()
			Expect(<-errors).To(Equal(context.Canceled))
		})

		Context("service type enumeration", func() {
			req := &dns.Msg{}
			req.SetQuestion(
				AbsoluteTypeEnumerationDomain("example.org"),
				dns.TypePTR,
			)

			It("responds to service type enumeration queries", func() {
				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_services._dns-sd._udp.example.org.	120	IN	PTR	_http._tcp.example.org.`,
					`_services._dns-sd._udp.example.org.	120	IN	PTR	_other._udp.example.org.`,
				)
			})

			It("does not include service types for which there are no remaining instances", func() {
				By("removing one of the two _http._tcp instances")

				changed, err := server.Unadvertise(ctx, instanceA)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(changed).To(BeTrue())

				By("asserting that the _http._tcp service type is still included in the response")

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_services._dns-sd._udp.example.org.	120	IN	PTR	_http._tcp.example.org.`,
					`_services._dns-sd._udp.example.org.	120	IN	PTR	_other._udp.example.org.`,
				)

				By("removing the last remaining _http._tcp instance")

				changed, err = server.Unadvertise(ctx, instanceB)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(changed).To(BeTrue())

				By("asserting that the _http._tcp service type is no longer included in the response")

				res, _, err = client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_services._dns-sd._udp.example.org.	120	IN	PTR	_other._udp.example.org.`,
				)
			})
		})

		Context("service instance enumeration (aka browsing)", func() {
			req := &dns.Msg{}
			req.SetQuestion(
				AbsoluteInstanceEnumerationDomain("_http._tcp", "example.org"),
				dns.TypePTR,
			)

			It("responds to service instance enumeration queries", func() {
				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_http._tcp.example.org.	120	IN	PTR	Instance\ A._http._tcp.example.org.`,
					`_http._tcp.example.org.	120	IN	PTR	Instance\ B._http._tcp.example.org.`,
				)
			})

			It("does not include service instances that have been removed", func() {
				changed, err := server.Unadvertise(ctx, instanceA)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(changed).To(BeTrue())

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_http._tcp.example.org.	120	IN	PTR	Instance\ B._http._tcp.example.org.`,
				)
			})
		})

		Context("selective instance enumeration", func() {
			req := &dns.Msg{}
			req.SetQuestion(
				AbsoluteSelectiveInstanceEnumerationDomain("_printer", "_http._tcp", "example.org"),
				dns.TypePTR,
			)

			It("responds to selective service instance enumeration queries", func() {
				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_printer._sub._http._tcp.example.org.	120	IN	PTR	Instance\ A._http._tcp.example.org.`,
				)
			})

			It("does not include service instances that have been removed", func() {
				changed, err := server.Unadvertise(ctx, instanceA)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(changed).To(BeTrue())

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					// none
				)
			})
		})

		Context("instance 'lookup' queries", func() {
			req := &dns.Msg{}
			req.SetQuestion(
				AbsoluteServiceInstanceName("Instance A", "_http._tcp", "example.org"),
				dns.TypeANY,
			)

			It("responds to instance lookup queries", func() {
				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`Instance\ A._http._tcp.example.org.	120	IN	SRV	10 20 12345 a.example.com.`,
					`Instance\ A._http._tcp.example.org.	120	IN	TXT	"<key>=<instance-a>"`,
				)
			})

			It("responds to instance lookup queries for a specific record type", func() {
				req.SetQuestion(
					AbsoluteServiceInstanceName("Instance A", "_http._tcp", "example.org"),
					dns.TypeSRV,
				)

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`Instance\ A._http._tcp.example.org.	120	IN	SRV	10 20 12345 a.example.com.`,
				)
			})

			It("does not include service instances that have been removed", func() {
				changed, err := server.Unadvertise(ctx, instanceA)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(changed).To(BeTrue())

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					// none
				)
			})
		})

		Context("address (IP lookup) queries", func() {
			req := &dns.Msg{}
			req.SetQuestion(
				"b.example.com.",
				dns.TypeANY,
			)

			It("responds to address lookup queries", func() {
				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`b.example.com.	120	IN	A	192.168.20.1`,
					"b.example.com.	120	IN	AAAA	fe80::1ce5:3c8b:36f:53cf",
				)
			})

			It("does not include service instances that have been removed", func() {
				_, err := server.Unadvertise(ctx, instanceB)
				Expect(err).ShouldNot(HaveOccurred())

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					// none
				)
			})
		})

		Context("queries with a question class other than INET", func() {
			req := &dns.Msg{}
			req.SetQuestion(
				"b.example.com.",
				dns.TypeANY,
			)

			It("responds normally if the class ANY", func() {
				req.Question[0].Qclass = dns.ClassANY

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`b.example.com.	120	IN	A	192.168.20.1`,
					"b.example.com.	120	IN	AAAA	fe80::1ce5:3c8b:36f:53cf",
				)
			})

			It("responds with a non-existant domain error if the class is any other class", func() {
				req.Question[0].Qclass = dns.ClassCHAOS

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				Expect(res.Rcode).To(Equal(dns.RcodeNameError))
			})
		})
	})

	Describe("func Advertise()", func() {
		It("returns false if the service is already advertised", func() {
			changed, err := server.Advertise(
				ctx,
				instanceC,
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(changed).To(BeFalse())
		})

		It("returns true if an already-advertised service's records are modified", func() {
			instanceC.TargetHost = "different-c.example.com"

			changed, err := server.Advertise(
				ctx,
				instanceC,
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(changed).To(BeTrue())
		})
	})

	Describe("func Unadvertise()", func() {
		It("returns true if the service is advertised with different records", func() {
			instanceC.TargetHost = "different-c.example.com"

			changed, err := server.Unadvertise(ctx, instanceC)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(changed).To(BeTrue())
		})

		It("returns false if the service is not advertised", func() {
			changed, err := server.Unadvertise(ctx, instanceC)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(changed).To(BeTrue())

			changed, err = server.Unadvertise(ctx, instanceC)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(changed).To(BeFalse())
		})
	})

	Describe("func Run()", func() {
		It("exits when the context is canceled", func() {
			errors := make(chan error, 1)

			go func() {
				errors <- server.Run(ctx, "udp", "127.0.0.1:65353")
			}()

			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()

			var err error
			Eventually(errors).Should(Receive(&err))
			Expect(err).To(Equal(context.Canceled))
		})
	})
})

func expectRecords(res *dns.Msg, records ...string) {
	var actual []string

	for _, rr := range res.Answer {
		actual = append(actual, rr.String())
	}

	Expect(actual).To(ConsistOf(records))
}
