package dnssd_test

import (
	. "github.com/dogmatiq/dissolve/dnssd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Instance", func() {
	var instance Instance

	BeforeEach(func() {
		instance = Instance{
			Name:    "Living Room TV.",
			Service: "_airplay._tcp",
			Domain:  "local",
		}
	})

	Describe("func FullyQualifiedName()", func() {
		It("returns the fully-qualified name, with appropriate escaping", func() {
			n := instance.FullyQualifiedName()
			Expect(n).To(Equal(`Living Room TV\.._airplay._tcp.local`))
		})
	})
})

var _ = Describe("func EscapeInstanceName()", func() {
	It("escapes dots and backslashes by adding a backslash", func() {
		n := EscapeInstanceName(`Foo\Bar.`)
		Expect(n).To(Equal(`Foo\\Bar\.`))
	})
})

var _ = Describe("func ParseInstanceName()", func() {
	It("unescapes dots and backslashes by removing the leading backslash", func() {
		n, tail, err := ParseInstanceName(`Foo\\Bar\.`)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(n).To(Equal(`Foo\Bar.`))
		Expect(tail).To(BeEmpty())
	})

	It("stops after the first DNS label", func() {
		n, tail, err := ParseInstanceName(`Foo\\Bar\..example.org`)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(n).To(Equal(`Foo\Bar.`))
		Expect(tail).To(Equal("example.org"))
	})

	It("returns an error if the name ends with the escape sequence", func() {
		_, _, err := ParseInstanceName(`Foo\`)
		Expect(err).To(MatchError("name is terminated with an escape character"))
	})
})
