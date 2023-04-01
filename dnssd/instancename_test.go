package dnssd_test

import (
	. "github.com/dogmatiq/dissolve/dnssd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("type ServiceInstanceName", func() {
	Describe("func Absolute()", func() {
		It("returns the absolute name of the service instance", func() {
			n := ServiceInstanceName{
				Name:        "Boardroom Printer",
				ServiceType: "_http._tcp",
				Domain:      "example.org",
			}

			Expect(n.Absolute()).To(Equal(`Boardroom\ Printer._http._tcp.example.org.`))
		})
	})

	Describe("func Relative()", func() {
		It("returns the relative name of the service instance", func() {
			n := ServiceInstanceName{
				Name:        "Boardroom Printer",
				ServiceType: "_http._tcp",
				Domain:      "example.org",
			}

			Expect(n.Relative()).To(Equal(`Boardroom\ Printer._http._tcp`))
		})
	})

	Describe("func Equal()", func() {
		DescribeTable(
			"it returns true if the names are equal",
			func(a, b ServiceInstanceName) {
				Expect(a.Equal(b)).To(BeTrue())
			},
			Entry(
				"zero-value",
				ServiceInstanceName{},
				ServiceInstanceName{},
			),
			Entry(
				"fully populated",
				ServiceInstanceName{
					Name:        "Boardroom Printer",
					ServiceType: "_http._tcp",
					Domain:      "example.org",
				},
				ServiceInstanceName{
					Name:        "Boardroom Printer",
					ServiceType: "_http._tcp",
					Domain:      "example.org",
				},
			),
		)

		DescribeTable(
			"it returns false if the names are not equal",
			func(a, b ServiceInstanceName) {
				Expect(a.Equal(b)).To(BeFalse())
			},
			Entry(
				"different name",
				ServiceInstanceName{
					Name: "Boardroom Printer",
				},
				ServiceInstanceName{
					Name: "Boardroom Printer 2",
				},
			),
			Entry(
				"different service type",
				ServiceInstanceName{
					ServiceType: "_http._tcp",
				},
				ServiceInstanceName{
					ServiceType: "_other._udp",
				},
			),
			Entry(
				"different domain",
				ServiceInstanceName{
					Domain: "example.org",
				},
				ServiceInstanceName{
					Domain: "example.com",
				},
			),
		)
	})
})

var _ = Describe("func AbsoluteServiceInstanceName()", func() {
	It("returns the fully-qualified name, with appropriate escaping", func() {
		d := AbsoluteServiceInstanceName("Boardroom Printer.", "_http._tcp", "example.org")
		Expect(d).To(Equal(`Boardroom\ Printer\.._http._tcp.example.org.`))
	})
})

var _ = Describe("func RelativeServiceInstanceName()", func() {
	It("returns the relative name, with appropriate escaping", func() {
		d := RelativeServiceInstanceName("Boardroom Printer.", "_http._tcp")
		Expect(d).To(Equal(`Boardroom\ Printer\.._http._tcp`))
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
