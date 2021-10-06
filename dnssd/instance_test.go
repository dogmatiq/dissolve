package dnssd_test

import (
	. "github.com/dogmatiq/dissolve/dnssd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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
