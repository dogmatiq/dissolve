package dnssd

import (
	"time"
)

// ServiceInstance encapsulates information about a single DNS-SD service
// instance.
type ServiceInstance struct {
	ServiceInstanceName

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
	Attributes AttributeCollection

	// TTL is the time-to-live of the instance's DNS records.
	TTL time.Duration
}

// Equal returns true if i and inst are equal.
func (i ServiceInstance) Equal(inst ServiceInstance) bool {
	return i.ServiceInstanceName.Equal(inst.ServiceInstanceName) &&
		i.TargetHost == inst.TargetHost &&
		i.TargetPort == inst.TargetPort &&
		i.Priority == inst.Priority &&
		i.Weight == inst.Weight &&
		i.Attributes.Equal(inst.Attributes) &&
		i.TTL == inst.TTL
}
