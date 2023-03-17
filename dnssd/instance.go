package dnssd

import (
	"errors"
	"strings"
	"time"
)

// ServiceInstance encapsulates information about a single DNS-SD service
// instance.
type ServiceInstance struct {
	// Name is the service instance's unqualified name.
	//
	// For example, "Boardroom Printer".
	//
	// This is the "<instance>" portion of the "service instance name", as
	// described in https://www.rfc-editor.org/rfc/rfc6763#section-4.1.
	Name string

	// ServiceType is the type of service that the instance provides.
	//
	// For example "_http._tcp", or "_airplay._tcp".
	//
	// This is the "<service>" portion of the "service instance name", as
	// described in https://www.rfc-editor.org/rfc/rfc6763#section-4.1.
	ServiceType string

	// Domain is the domain under which the instance is advertised.
	//
	// That is, the domain name that contains the DNS-SD records SRV, PTR and
	// TXT records.
	//
	// This name is often set to "local" when using Multicast DNS (Bonjour,
	// Zerconf), but may be any valid domain name.
	//
	// This is the "<domain>" portion of the "service instance name", as
	// described in https://www.rfc-editor.org/rfc/rfc6763#section-4.1.
	Domain string

	// TargetHost is the fully-qualified hostname of the machine that hosts the
	// service.
	//
	// This is not necessarily within in the same domain as the DNS-SD records.
	TargetHost string

	// TargetPort is TCP or UDP port on which the service is provided.
	TargetPort uint16

	// Priority is the priority of the instance within the pool of instances
	// that  offer the same service for the same domain.
	//
	// It controls which servers are contacted first. Lower values have a higher
	// priority.
	//
	// See https://www.rfc-editor.org/rfc/rfc2782.
	Priority uint16

	// Weight is the weight of this instance within the pool of instances that
	// offer the same service for the same domain.
	//
	// It controls the likelihood that the instance will be chosen from a pool
	// instances with the same priority. Higher values are more likely to be
	// chosen.
	//
	// See https://www.rfc-editor.org/rfc/rfc2782.
	Weight uint16

	// Attributes contains a set of attributes that provide additional
	// information about the service instance.
	//
	// Attributes are encoded in the instance's TXT record, as per
	// https://www.rfc-editor.org/rfc/rfc6763#section-6.3.
	//
	// Each element in the slice corresponds to the attributes encoded in a
	// single TXT record.
	//
	// Each instance MUST have at least one TXT record and MAY have multiple TXT
	// records, although this is rarely used in practice, see
	// https://www.rfc-editor.org/rfc/rfc6763#section-6.8.
	//
	// An empty slice is a valid representation of an instance with a single
	// empty TXT record.
	Attributes []Attributes

	// TTL is the time-to-live of the instance's DNS records.
	TTL time.Duration
}

// ServiceInstanceName returns the fully-qualfied DNS domain name that is
// queried to lookup records about a single service instance.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-4.1.
func ServiceInstanceName(instance, serviceType, domain string) string {
	return EscapeInstance(instance) + "." + InstanceEnumerationDomain(serviceType, domain)
}

// needsEscape is a string containing runes that must be escaped when they
// appear in an instance name.
const needsEscape = `. '@;()"\`

// EscapeInstance escapes a service instance name for use within DNS
// records.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-4.3.
func EscapeInstance(instance string) string {
	// https://www.rfc-editor.org/rfc/rfc6763#section-4.3
	//
	// This document RECOMMENDS that if concatenating the three portions of
	// a Service Instance Name, any dots in the <Instance> portion be
	// escaped following the customary DNS convention for text files: by
	// preceding literal dots with a backslash (so "." becomes "\.").
	// Likewise, any backslashes in the <Instance> portion should also be
	// escaped by preceding them with a backslash (so "\" becomes "\\").

	var w strings.Builder

	for _, r := range instance {
		if strings.ContainsRune(needsEscape, r) {
			w.WriteRune('\\')
		}

		w.WriteRune(r)
	}

	return w.String()
}

// ParseInstance parses the "<instance>" portion of a service instance name.
//
// The given name must be either an escaped "<instance>" portion of a
// fully-qualified "service instance name", or the fully-qualified "service
// instance name" itself. Parsing stops at the first unescaped dot.
//
// instance is the parsed and unescaped instance name. tail is the remaining
// unparsed portion of n, not including the separating dot.
//
// tail is empty if name is just the "<instance>" portion (that is, it does not
// contain any unescaped dots).
func ParseInstance(name string) (instance, tail string, err error) {
	// https://www.rfc-editor.org/rfc/rfc6763#section-4.3
	//
	// This document RECOMMENDS that if concatenating the three portions of
	// a Service Instance Name, any dots in the <Instance> portion be
	// escaped following the customary DNS convention for text files: by
	// preceding literal dots with a backslash (so "." becomes "\.").
	// Likewise, any backslashes in the <Instance> portion should also be
	// escaped by preceding them with a backslash (so "\" becomes "\\").
	var w strings.Builder
	escaped := false

	for i, r := range name {
		if escaped {
			escaped = false
		} else if r == '\\' {
			escaped = true
			continue
		} else if r == '.' {
			tail = name[i+1:] // we know '.' is a single byte
			break
		}

		w.WriteRune(r)
	}

	if escaped {
		return "", "", errors.New("name is terminated with an escape character")
	}

	return w.String(), tail, nil
}
