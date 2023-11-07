package dnssd_test

import (
	"time"

	. "github.com/dogmatiq/dissolve/dnssd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("type ServiceInstance", func() {
	Describe("func Equal()", func() {
		DescribeTable(
			"it returns true if the instances are equal",
			func(a, b ServiceInstance) {
				Expect(a.Equal(b)).To(BeTrue())
			},
			Entry(
				"zero-value",
				ServiceInstance{},
				ServiceInstance{},
			),
			Entry(
				"fully populated",
				ServiceInstance{
					ServiceInstanceName: ServiceInstanceName{
						Name:        "Boardroom Printer",
						ServiceType: "_http._tcp",
						Domain:      "example.org",
					},
					TargetHost: "boardroom-printer.example.org",
					TargetPort: 80,
					Priority:   10,
					Weight:     20,
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{1}).
							WithFlag("default"),
					},
					TTL: 30 * time.Second,
				},
				ServiceInstance{
					ServiceInstanceName: ServiceInstanceName{
						Name:        "Boardroom Printer",
						ServiceType: "_http._tcp",
						Domain:      "example.org",
					},
					TargetHost: "boardroom-printer.example.org",
					TargetPort: 80,
					Priority:   10,
					Weight:     20,
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{1}).
							WithFlag("default"),
					},
					TTL: 30 * time.Second,
				},
			),
			Entry(
				"attributes in a different order",
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{1}).
							WithFlag("default"),
						NewAttributes().
							WithPair("txtvers", []byte{2}),
					},
				},
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{2}),
						NewAttributes().
							WithFlag("default").
							WithPair("txtvers", []byte{1}),
					},
				},
			),
			Entry(
				"multiple copies of the same set of attributes",
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{1}),
						NewAttributes().
							WithPair("txtvers", []byte{1}),
					},
				},
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{1}),
						NewAttributes().
							WithPair("txtvers", []byte{1}),
					},
				},
			),
		)
		DescribeTable(
			"it returns false if the instances are not equal",
			func(a, b ServiceInstance) {
				Expect(a.Equal(b)).To(BeFalse())
			},
			Entry(
				"different name",
				ServiceInstance{
					ServiceInstanceName: ServiceInstanceName{
						Name: "Boardroom Printer",
					},
				},
				ServiceInstance{
					ServiceInstanceName: ServiceInstanceName{
						Name: "Boardroom Printer 2",
					},
				},
			),
			Entry(
				"different service type",
				ServiceInstance{
					ServiceInstanceName: ServiceInstanceName{
						ServiceType: "_http._tcp",
					},
				},
				ServiceInstance{
					ServiceInstanceName: ServiceInstanceName{
						ServiceType: "_other._udp",
					},
				},
			),
			Entry(
				"different domain",
				ServiceInstance{
					ServiceInstanceName: ServiceInstanceName{
						Domain: "example.org",
					},
				},
				ServiceInstance{
					ServiceInstanceName: ServiceInstanceName{
						Domain: "example.com",
					},
				},
			),
			Entry(
				"different target host",
				ServiceInstance{TargetHost: "boardroom-printer.example.org"},
				ServiceInstance{TargetHost: "boardroom-printer-2.example.org"},
			),
			Entry(
				"different target port",
				ServiceInstance{TargetPort: 80},
				ServiceInstance{TargetPort: 8080},
			),
			Entry(
				"different priority",
				ServiceInstance{Priority: 10},
				ServiceInstance{Priority: 20},
			),
			Entry(
				"different weight",
				ServiceInstance{Weight: 20},
				ServiceInstance{Weight: 30},
			),
			Entry(
				"different TTL",
				ServiceInstance{TTL: 30 * time.Second},
				ServiceInstance{TTL: 60 * time.Second},
			),
			Entry(
				"different attributes",
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{1}),
					},
				},
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{2}),
					},
				},
			),
			Entry(
				"different attributes - flag vs empty value",
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("flag", nil),
					},
				},
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithFlag("flag"),
					},
				},
			),
			Entry(
				"different attributes - subset",
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{1}),
					},
				},
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{1}), // <-- same
						NewAttributes().
							WithPair("txtvers", []byte{2}), // <-- different
					},
				},
			),
			Entry(
				"different attributes - multiple copies of the same set of attributes",
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{1}), // <-- same
						NewAttributes().
							WithPair("txtvers", []byte{1}), // <-- same
					},
				},
				ServiceInstance{
					Attributes: AttributeCollection{
						NewAttributes().
							WithPair("txtvers", []byte{1}), // <-- same
						NewAttributes().
							WithPair("txtvers", []byte{2}), // <-- different
					},
				},
			),
		)
	})
})
