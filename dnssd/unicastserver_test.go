package dnssd_test

import (
	"context"
	"net"
	"time"

	. "github.com/dogmatiq/dissolve/dnssd"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
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
			Instance:    "Instance A",
			ServiceType: "_http._tcp",
			Domain:      "local",
			TargetHost:  "a.example.org",
			TargetPort:  12345,
			Priority:    10,
			Weight:      20,
			Attributes:  []Attributes{{}},
		}
		instanceA.Attributes[0].Set("<key>", []byte("<instance-a>"))

		instanceB = ServiceInstance{
			Instance:    "Instance B",
			ServiceType: "_http._tcp",
			Domain:      "local",
			TargetHost:  "b.example.org",
			TargetPort:  12345,
			Priority:    10,
			Weight:      20,
			Attributes:  []Attributes{{}, {}},
		}
		instanceB.Attributes[0].Set("<key>", []byte("<instance-b0>"))
		instanceB.Attributes[1].Set("<key>", []byte("<instance-b1>"))

		instanceC = ServiceInstance{
			Instance:    "Instance C",
			ServiceType: "_sleep-proxy._udp",
			Domain:      "local",
			TargetHost:  "c.example.org",
			TargetPort:  12345,
			Priority:    10,
			Weight:      20,
		}

		server = &UnicastServer{}

		server.Advertise(
			instanceA,
			WithServiceSubType("_printer"),
		)

		server.Advertise(
			instanceB,
			WithIPAddress(net.IPv4(192, 168, 20, 1)),
			WithIPAddress(net.ParseIP("fe80::1ce5:3c8b:36f:53cf")),
		)

		server.Advertise(instanceC)
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
				TypeEnumerationDomain("local")+".",
				dns.TypePTR,
			)

			It("responds to service type enumeration queries", func() {
				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_services._dns-sd._udp.local.	120	IN	PTR	_http._tcp.local.`,
					`_services._dns-sd._udp.local.	120	IN	PTR	_sleep-proxy._udp.local.`,
				)
			})

			It("does not include service types for which there are no remaining instances", func() {
				By("removing one of the two _http._tcp instances")

				server.Remove(instanceA)

				By("asserting that the _http._tcp service type is still included in the response")

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_services._dns-sd._udp.local.	120	IN	PTR	_http._tcp.local.`,
					`_services._dns-sd._udp.local.	120	IN	PTR	_sleep-proxy._udp.local.`,
				)

				By("removing the last remaining _http._tcp instance")

				server.Remove(instanceB)

				By("asserting that the _http._tcp service type is no longer included in the response")

				res, _, err = client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_services._dns-sd._udp.local.	120	IN	PTR	_sleep-proxy._udp.local.`,
				)
			})
		})

		Context("service instance enumeration (aka browsing)", func() {
			req := &dns.Msg{}
			req.SetQuestion(
				InstanceEnumerationDomain("_http._tcp", "local")+".",
				dns.TypePTR,
			)

			It("responds to service instance enumeration queries", func() {
				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_http._tcp.local.	120	IN	PTR	Instance\ A._http._tcp.local.`,
					`_http._tcp.local.	120	IN	PTR	Instance\ B._http._tcp.local.`,
				)
			})

			It("does not include service instances that have been removed", func() {
				server.Remove(instanceA)

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_http._tcp.local.	120	IN	PTR	Instance\ B._http._tcp.local.`,
				)
			})
		})

		Context("selective instance enumeration", func() {
			req := &dns.Msg{}
			req.SetQuestion(
				SelectiveInstanceEnumerationDomain("_printer", "_http._tcp", "local")+".",
				dns.TypePTR,
			)

			It("responds to selective service instance enumeration queries", func() {
				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`_printer._sub._http._tcp.local.	120	IN	PTR	Instance\ A._http._tcp.local.`,
				)
			})

			It("does not include service instances that have been removed", func() {
				server.Remove(instanceA)

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
				ServiceInstanceName("Instance A", "_http._tcp", "local")+".",
				dns.TypeANY,
			)

			It("responds to instance lookup queries", func() {
				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`Instance\ A._http._tcp.local.	120	IN	SRV	10 20 12345 a.example.org.`,
					`Instance\ A._http._tcp.local.	120	IN	TXT	"<key>=<instance-a>"`,
				)
			})

			It("responds to instance lookup queries for a specific record type", func() {
				req.SetQuestion(
					ServiceInstanceName("Instance A", "_http._tcp", "local")+".",
					dns.TypeSRV,
				)

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`Instance\ A._http._tcp.local.	120	IN	SRV	10 20 12345 a.example.org.`,
				)
			})

			It("does not include service instances that have been removed", func() {
				server.Remove(instanceA)

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
				"b.example.org.",
				dns.TypeANY,
			)

			It("responds to address lookup queries", func() {
				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					`b.example.org.	120	IN	A	192.168.20.1`,
					"b.example.org.	120	IN	AAAA	fe80::1ce5:3c8b:36f:53cf",
				)
			})

			It("does not include service instances that have been removed", func() {
				server.Remove(instanceB)

				res, _, err := client.ExchangeContext(ctx, req, "127.0.0.1:65353")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).NotTo(BeNil())
				expectRecords(
					res,
					// none
				)
			})
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
				time.Sleep(50 * time.Millisecond)
				close(errors)
			}()

			Expect(<-errors).To(Equal(context.Canceled))
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
