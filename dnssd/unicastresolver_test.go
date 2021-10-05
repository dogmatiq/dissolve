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
		}
		instanceA.Attributes.Set("<key>", []byte("<instance-a>"))

		instanceB = ServiceInstance{
			Instance:    "Instance B",
			ServiceType: "_http._tcp",
			Domain:      "local",
			TargetHost:  "b.example.org",
			TargetPort:  12345,
			Priority:    10,
			Weight:      20,
		}
		instanceB.Attributes.Set("<key>", []byte("<instance-b>"))

		instanceC = ServiceInstance{
			Instance:    "Instance C",
			ServiceType: "_sleep-proxy._udp",
			Domain:      "local",
			TargetHost:  "c.example.org",
			TargetPort:  12345,
			Priority:    10,
			Weight:      20,
		}
		instanceC.Attributes.Set("<key>", []byte("<instance-c>"))

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
})
