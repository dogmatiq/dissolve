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
					Name:        "Boardroom Printer",
					ServiceType: "_http._tcp",
					Domain:      "example.org",
					TargetHost:  "boardroom-printer.example.org",
					TargetPort:  80,
					Priority:    10,
					Weight:      20,
					Attributes: []Attributes{
						NewAttributes().
							WithPair("txtvers", []byte{1}).
							WithFlag("default"),
					},
					TTL: 30 * time.Second,
				},
				ServiceInstance{
					Name:        "Boardroom Printer",
					ServiceType: "_http._tcp",
					Domain:      "example.org",
					TargetHost:  "boardroom-printer.example.org",
					TargetPort:  80,
					Priority:    10,
					Weight:      20,
					Attributes: []Attributes{
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
					Attributes: []Attributes{
						NewAttributes().
							WithPair("txtvers", []byte{1}).
							WithFlag("default"),
						NewAttributes().
							WithPair("txtvers", []byte{2}),
					},
				},
				ServiceInstance{
					Attributes: []Attributes{
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
					Attributes: []Attributes{
						NewAttributes().
							WithPair("txtvers", []byte{1}),
						NewAttributes().
							WithPair("txtvers", []byte{1}),
					},
				},
				ServiceInstance{
					Attributes: []Attributes{
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
				ServiceInstance{Name: "Boardroom Printer"},
				ServiceInstance{Name: "Boardroom Printer 2"},
			),
			Entry(
				"different service type",
				ServiceInstance{ServiceType: "_http._tcp"},
				ServiceInstance{ServiceType: "_other._udp"},
			),
			Entry(
				"different domain",
				ServiceInstance{Domain: "example.org"},
				ServiceInstance{Domain: "example.com"},
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
					Attributes: []Attributes{
						NewAttributes().
							WithPair("txtvers", []byte{1}),
					},
				},
				ServiceInstance{
					Attributes: []Attributes{
						NewAttributes().
							WithPair("txtvers", []byte{2}),
					},
				},
			),
			Entry(
				"different attributes - flag vs empty value",
				ServiceInstance{
					Attributes: []Attributes{
						NewAttributes().
							WithPair("flag", nil),
					},
				},
				ServiceInstance{
					Attributes: []Attributes{
						NewAttributes().
							WithFlag("flag"),
					},
				},
			),
			Entry(
				"different attributes - subset",
				ServiceInstance{
					Attributes: []Attributes{
						NewAttributes().
							WithPair("txtvers", []byte{1}),
					},
				},
				ServiceInstance{
					Attributes: []Attributes{
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
					Attributes: []Attributes{
						NewAttributes().
							WithPair("txtvers", []byte{1}), // <-- same
						NewAttributes().
							WithPair("txtvers", []byte{1}), // <-- same
					},
				},
				ServiceInstance{
					Attributes: []Attributes{
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

var _ = Describe("func ServiceInstanceName()", func() {
	It("returns the fully-qualified name, with appropriate escaping", func() {
		d := ServiceInstanceName("Boardroom Printer.", "_http._tcp", "example.org")
		Expect(d).To(Equal(`Boardroom\ Printer\.._http._tcp.example.org`))
	})
})

var _ = Describe("func EscapeInstance()", func() {
	It("escapes special characters by adding a backslash", func() {
		n := EscapeInstance(`. '@;()"\regulartext`)
		Expect(n).To(Equal(`\.\ \'\@\;\(\)\"\\regulartext`))
	})
})

var _ = Describe("func ParseInstance()", func() {
	It("unescapes dots and backslashes by removing the leading backslash", func() {
		n, tail, err := ParseInstance(`Foo\\Bar\.`)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(n).To(Equal(`Foo\Bar.`))
		Expect(tail).To(BeEmpty())
	})

	It("stops after the first DNS label", func() {
		n, tail, err := ParseInstance(`Foo\\Bar\..example.org`)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(n).To(Equal(`Foo\Bar.`))
		Expect(tail).To(Equal("example.org"))
	})

	It("returns an error if the name ends with the escape sequence", func() {
		_, _, err := ParseInstance(`Foo\`)
		Expect(err).To(MatchError("name is terminated with an escape character"))
	})
})
