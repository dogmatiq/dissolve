package dnssd_test

import (
	. "github.com/dogmatiq/dissolve/dnssd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func TypeEnumerationDomain()", func() {
	It("returns the 'type enumeration domain' for the given domain", func() {
		d := TypeEnumerationDomain("example.org")
		Expect(d).To(Equal("_services._dns-sd._udp.example.org"))
	})
})

var _ = Describe("func InstanceEnumerationDomain()", func() {
	It("returns the 'instance enumeration domain' for the given service & domain", func() {
		d := InstanceEnumerationDomain("_http._tcp", "example.org")
		Expect(d).To(Equal("_http._tcp.example.org"))
	})
})

var _ = Describe("func SelectiveInstanceEnumerationDomain()", func() {
	It("returns the 'instance enumeration domain' for the given service & domain", func() {
		d := SelectiveInstanceEnumerationDomain("_printer", "_http._tcp", "example.org")
		Expect(d).To(Equal("_printer._sub._http._tcp.example.org"))
	})
})
