package dnssd

import (
	"errors"
	"strings"
	"time"
)

// Instance is a DNS-SD service instance.
type Instance struct {
	// Name is the service instance's unqualified name. That is, without the
	// service or domain components.
	Name string

	// Service is the type of service that the instance provides.
	//
	// For example "_http._tcp", or "_airplay._tcp".
	//
	// See https://datatracker.ietf.org/doc/html/rfc6763#section-4.1.2.
	Service string

	// Domain is the domain under which the instance is advertised.
	//
	// That is, the domain name that contains the DNS-SD records SRV, PTR and
	// TXT records.
	//
	// See https://datatracker.ietf.org/doc/html/rfc6763#section-4.1.3
	Domain string

	// TargetHost is the fully-qualified hostname of the machine that hosts the
	// service.
	//
	// This is not necessarily within in the same domain as the DNS-SD records.
	TargetHost string

	// TargetPort is TCP or UDP port on which the service is provided.
	TargetPort uint16

	// Attributes contains a set of attributes that provide additional
	// information about the service instance.
	//
	// Attributes are encoded in the instance's TXT record, as per
	// https://datatracker.ietf.org/doc/html/rfc6763#section-6.3.
	Attributes Attributes

	// Priority is the priority of the instance within the pool of instances
	// that  offer the same service for the same domain.
	//
	// It controls which servers are contacted first. Lower values have a higher
	// priority.
	//
	// See https://datatracker.ietf.org/doc/html/rfc2782.
	Priority uint16

	// Weight is the weight of this instance within the pool of instances that
	// offer the same service for the same domain.
	//
	// It controls the likelihood that the instance will be chosen from a pool
	// instances with the same priority. Higher values are more likely to be
	// chosen.
	//
	// See https://datatracker.ietf.org/doc/html/rfc2782.
	Weight uint16

	// TTL is the time-to-live of the instance's DNS records.
	TTL time.Duration
}

// FullyQualifiedName returns the fully-qualified instance name of this
// instance, including the name, service and domain components.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-4.1 for a
// description of how fully-qualified service names are structured.
func (i Instance) FullyQualifiedName() string {
	return EscapeInstanceName(i.Name) + "." + InstanceEnumDomain(i.Service, i.Domain)
}

// EscapeInstanceName escapes a service instance name for use within DNS
// records.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-4.3
func EscapeInstanceName(n string) string {
	// https://datatracker.ietf.org/doc/html/rfc6763#section-4.3
	//
	// This document RECOMMENDS that if concatenating the three portions of
	// a Service Instance Name, any dots in the <Instance> portion be
	// escaped following the customary DNS convention for text files: by
	// preceding literal dots with a backslash (so "." becomes "\.").
	// Likewise, any backslashes in the <Instance> portion should also be
	// escaped by preceding them with a backslash (so "\" becomes "\\").

	var w strings.Builder

	for {
		i := strings.IndexAny(n, `.\`)
		if i == -1 {
			w.WriteString(n)
			break
		}

		w.WriteString(n[:i])
		w.WriteByte('\\')
		w.WriteByte(n[i])
		n = n[i+1:]
	}

	return w.String()
}

// ParseInstanceName parses a service instance name.
//
// n must be an escaped service instance name, either relative or fully
// qualified. Parsing stops at the first unescaped dot.
//
// name is the parsed and unescaped instance name. tail is the remaining
// unparsed portion of n, not including the separating dot.
//
// If n contains only a single label (that is, does not contain any unescaped
// dots), tail is empty.
func ParseInstanceName(n string) (name, tail string, err error) {
	var w strings.Builder

	// https://tools.ietf.org/html/rfc6763#section-4.3
	//
	// This document RECOMMENDS that if concatenating the three portions of
	// a Service Instance Name, any dots in the <Instance> portion be
	// escaped following the customary DNS convention for text files: by
	// preceding literal dots with a backslash (so "." becomes "\.").
	// Likewise, any backslashes in the <Instance> portion should also be
	// escaped by preceding them with a backslash (so "\" becomes "\\").
	escaped := false

	// Iterate over bytes (not runes, hence not using range keyword).
	for i := 0; i < len(n); i++ {
		ch := n[i]

		if escaped {
			escaped = false
		} else if ch == '\\' {
			escaped = true
			continue
		} else if ch == '.' {
			tail = n[i+1:]
			break
		}

		w.WriteByte(ch)
	}

	if escaped {
		return "", "", errors.New("name is terminated with an escape character")
	}

	return w.String(), tail, nil
}
