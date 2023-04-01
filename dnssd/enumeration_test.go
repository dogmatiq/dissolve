package dnssd_test

import (
	. "github.com/dogmatiq/dissolve/dnssd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("func AbsoluteTypeEnumerationDomain()", func() {
	It("returns the absolute 'type enumeration domain' for the given domain", func() {
		d := AbsoluteTypeEnumerationDomain("example.org")
		Expect(d).To(Equal("_services._dns-sd._udp.example.org."))
	})
})

var _ = Describe("func RelativeTypeEnumerationDomain()", func() {
	It("returns the relative 'type enumeration domain'", func() {
		d := RelativeTypeEnumerationDomain()
		Expect(d).To(Equal("_services._dns-sd._udp"))
	})
})

var _ = Describe("func AbsoluteInstanceEnumerationDomain()", func() {
	It("returns the absolute 'instance enumeration domain' for the given service type & domain", func() {
		d := AbsoluteInstanceEnumerationDomain("_http._tcp", "example.org")
		Expect(d).To(Equal("_http._tcp.example.org."))
	})
})

var _ = Describe("func RelativeInstanceEnumerationDomain()", func() {
	It("returns the relative 'instance enumeration domain' for the given service type", func() {
		d := RelativeInstanceEnumerationDomain("_http._tcp")
		Expect(d).To(Equal("_http._tcp"))
	})
})

var _ = Describe("func AbsoluteSelectiveInstanceEnumerationDomain()", func() {
	It("returns the absolute 'selective instance enumeration domain' for the given sub-type, service type & domain", func() {
		d := AbsoluteSelectiveInstanceEnumerationDomain("_printer", "_http._tcp", "example.org")
		Expect(d).To(Equal("_printer._sub._http._tcp.example.org."))
	})
})

var _ = Describe("func RelativeSelectiveInstanceEnumerationDomain()", func() {
	It("returns the relative 'selective instance enumeration domain' for the given sub-type, service type ", func() {
		d := RelativeSelectiveInstanceEnumerationDomain("_printer", "_http._tcp")
		Expect(d).To(Equal("_printer._sub._http._tcp"))
	})
})
