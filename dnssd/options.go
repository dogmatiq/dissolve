package dnssd

import "net"

// AdvertiseOption is an option that changes the behavior of how a service
// instance is advertised.
type AdvertiseOption func(*advertiseOptions)

// WithIPAddress is an AccountOption that adds DNS A and/or AAAA records that
// map the service's hostname to the given IP.
func WithIPAddress(ip net.IP) AdvertiseOption {
	return func(opts *advertiseOptions) {
		opts.IPAddresses = append(opts.IPAddresses, ip)
	}
}

// WithServiceSubType is an announce option that advertises the service as
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
