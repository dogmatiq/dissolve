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

var _ = Context("UnicastResolver", func() {
	var (
		ctx                             context.Context
		cancel                          context.CancelFunc
		instanceA, instanceB, instanceC ServiceInstance
		server                          *UnicastServer
		serverResult                    chan error
		resolver                        *UnicastResolver
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
			TTL:         DefaultTTL / 2, // Use something other than the default for LookupInstance() test
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

		serverResult = make(chan error, 1)

		go func() {
			serverResult <- server.Run(ctx, "udp", "127.0.0.1:65353")
		}()

		// Fudge-factor to allow the server time to start.
		time.Sleep(100 * time.Millisecond)

		resolver = &UnicastResolver{
			Config: &dns.ClientConfig{
				Servers: []string{"127.0.0.1"},
				Port:    "65353",
			},
		}
	})

	AfterEach(func() {
		cancel()
		Expect(<-serverResult).To(Equal(context.Canceled))
	})

	Describe("func EnumerateServiceTypes()", func() {
		It("returns the distinct service types advertised within the domain", func() {
			serviceTypes, err := resolver.EnumerateServiceTypes(ctx, "local")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(serviceTypes).To(ContainElements(
				"_http._tcp",
				"_sleep-proxy._udp",
			))
		})
	})

	Describe("func EnumerateInstances()", func() {
		It("returns instances of the service type that are advertised within the domain", func() {
			serviceTypes, err := resolver.EnumerateInstances(ctx, "_http._tcp", "local")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(serviceTypes).To(ContainElements(
				"Instance A",
				"Instance B",
			))
		})
	})

	Describe("func EnumerateInstancesBySubType()", func() {
		It("returns instances of the sub-type and service type that are advertised within the domain", func() {
			serviceTypes, err := resolver.EnumerateInstancesBySubType(ctx, "_printer", "_http._tcp", "local")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(serviceTypes).To(ContainElements(
				"Instance A",
			))
		})
	})

	Describe("func LookupServiceInstance()", func() {
		It("returns complete information about the service instance", func() {
			i, ok, err := resolver.LookupInstance(ctx, "Instance A", "_http._tcp", "local")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(i).To(Equal(instanceA))
		})

		It("returns false if no such instance exists", func() {
			_, ok, err := resolver.LookupInstance(ctx, "Instance X", "_http._tcp", "local")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeFalse())
		})
	})
})
