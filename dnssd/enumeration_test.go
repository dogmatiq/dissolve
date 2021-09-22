package dnssd_test

import (
	. "github.com/dogmatiq/dissolve/dnssd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func TypeEnumDomain()", func() {
	It("returns the 'type enumeration domain' for the given domain", func() {
		d := TypeEnumDomain("example.org")
		Expect(d).To(Equal("_services._dns-sd._udp.example.org"))
	})
})

var _ = Describe("func InstanceEnumDomain()", func() {
	It("returns the 'instance enumeration domain' for the given service & domain", func() {
		d := InstanceEnumDomain("_http._tcp", "example.org")
		Expect(d).To(Equal("_http._tcp.example.org"))
	})
})

var _ = Describe("func SelectiveInstanceEnumDomain()", func() {
	It("returns the 'instance enumeration domain' for the given service & domain", func() {
		d := SelectiveInstanceEnumDomain("_printer", "_http._tcp", "example.org")
		Expect(d).To(Equal("_printer._sub._http._tcp.example.org"))
	})
})
