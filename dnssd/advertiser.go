package dnssd

import (
	"context"
	"fmt"
	"net"
)

// Advertiser is an interface for advertising DNS-SD service via a unicast DNS
// provider.
type Advertiser interface {
	// Advertise creates and/or updates DNS records to advertise the given
	// service instance.
	//
	// It returns true if any changes to DNS records were made, or false if the
	// service was already advertised as-is.
	Advertise(
		ctx context.Context,
		inst ServiceInstance,
		options ...AdvertiseOption,
	) (bool, error)

	// Unadvertise removes and/or updates DNS records to stop advertising the
	// given service instance.
	//
	// It true if any changes to DNS records were made, or false if the service
	// was not advertised.
	Unadvertise(
		ctx context.Context,
		inst ServiceInstance,
	) (bool, error)
}

// UnsupportedDomainError is returned by the methods of an [Advertiser] when the
// advertiser does not support advertising on the service's domain name.
type UnsupportedDomainError struct {
	Domain string
	Cause  error
}

func (e UnsupportedDomainError) Error() string {
	message := fmt.Sprintf(
		"advertising DNS-SD service instances on the %q domain is not supported",
		e.Domain,
	)

	if e.Cause == nil {
		return message
	}

	return fmt.Sprintf("%s: %s", message, e.Cause)
}

// AdvertiseOption is an option that changes the behavior of how a service
// instance is advertised.
type AdvertiseOption func(*advertiseOptions)

// WithIPAddress is an [AdvertiseOption] that adds A and/or AAAA records that
// map the service's hostname to the given IP.
func WithIPAddress(ip net.IP) AdvertiseOption {
	return func(opts *advertiseOptions) {
		opts.IPAddresses = append(opts.IPAddresses, ip)
	}
}

// WithServiceSubType is an [AdvertiseOption] that advertises the service as
// providing a specific service sub-type.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-7.1
func WithServiceSubType(subType string) AdvertiseOption {
	return func(opts *advertiseOptions) {
		opts.ServiceSubTypes = append(opts.ServiceSubTypes, subType)
	}
}

type advertiseOptions struct {
	IPAddresses     []net.IP
	ServiceSubTypes []string
}

func resolveAdvertiseOptions(options []AdvertiseOption) advertiseOptions {
	var opts advertiseOptions

	for _, opt := range options {
		opt(&opts)
	}

	return opts
}
