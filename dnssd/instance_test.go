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
