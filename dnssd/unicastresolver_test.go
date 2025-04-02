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
			TTL: DefaultTTL / 2, // Use something other than the default for LookupInstance() test
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

		_, err := server.Advertise(
			ctx,
			instanceA,
			WithServiceSubType("_printer"),
		)
		Expect(err).ShouldNot(HaveOccurred())

		_, err = server.Advertise(
			ctx,
			instanceB,
			WithIPAddress(net.IPv4(192, 168, 20, 1)),
			WithIPAddress(net.ParseIP("fe80::1ce5:3c8b:36f:53cf")),
		)
		Expect(err).ShouldNot(HaveOccurred())

		_, err = server.Advertise(
			ctx,
			instanceC,
		)
		Expect(err).ShouldNot(HaveOccurred())

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
			serviceTypes, err := resolver.EnumerateServiceTypes(ctx, "example.org")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(serviceTypes).To(ContainElements(
				"_http._tcp",
				"_other._udp",
			))
		})
	})

	Describe("func EnumerateInstances()", func() {
		It("returns instances of the service type that are advertised within the domain", func() {
			serviceTypes, err := resolver.EnumerateInstances(ctx, "_http._tcp", "example.org")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(serviceTypes).To(ContainElements(
				"Instance A",
				"Instance B",
			))
		})
	})

	Describe("func EnumerateInstancesBySubType()", func() {
		It("returns instances of the sub-type and service type that are advertised within the domain", func() {
			serviceTypes, err := resolver.EnumerateInstancesBySubType(ctx, "_printer", "_http._tcp", "example.org")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(serviceTypes).To(ContainElements(
				"Instance A",
			))
		})
	})

	Describe("func LookupServiceInstance()", func() {
		It("returns complete information about the service instance", func() {
			i, ok, err := resolver.LookupInstance(ctx, "Instance A", "_http._tcp", "example.org")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(i).To(Equal(instanceA))
		})

		It("returns false if no such instance exists", func() {
			_, ok, err := resolver.LookupInstance(ctx, "Instance X", "_http._tcp", "example.org")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeFalse())
		})
	})
})
