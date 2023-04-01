package dnssd

import (
	"context"

	"github.com/dogmatiq/dissolve/internal/domainname"
)

// Enumerator is an interface for enumerating (discovering) DNS-SD services.
type Enumerator interface {
	// EnumerateServiceTypes finds all of the service types advertised within a
	// single domain.
	//
	// It blocks until ctx is canceled or an error occurs.
	//
	// obs is an observer fuction that is called whenever a new service type is
	// discovered. The context passed to obs is canceled when that service type
	// goes away. Enumeration is aborted if obs returns an error.
	EnumerateServiceTypes(
		ctx context.Context,
		domain string,
		obs func(ctx context.Context, serviceType string) error,
	) error

	// EnumerateInstances finds all of the instances of a specific service type
	// that are advertised within a single domain. This operation is also known
	// as "browsing".
	//
	// It blocks until ctx is canceled or an error occurs.
	//
	// obs is an observer fuction that is called whenever a new service instance
	// is discovered. The context passed to obs is canceled when that service
	// instance goes away. Enumeration is aborted if obs returns an error.
	EnumerateInstances(
		ctx context.Context,
		serviceType, domain string,
		obs func(ctx context.Context, i ServiceInstance) error,
	) error

	// EnumerateInstancesSelectively finds all of the instances of a specific
	// service type that are advertised within a single domain where those
	// services have a specific service sub-type.
	//
	// It blocks until ctx is canceled or an error occurs.
	//
	// obs is an observer fuction that is called whenever a new service instance
	// is discovered. The context passed to obs is canceled when that service
	// instance goes away. Enumeration is aborted if obs returns an error.
	EnumerateInstancesSelectively(
		ctx context.Context,
		subType, serviceType, domain string,
		obs func(ctx context.Context, i ServiceInstance) error,
	) error
}

// AbsoluteTypeEnumerationDomain returns the absolute DNS name that is queried
// to perform "service type enumeration" for a specific domain.
//
// Service type enumeration is used to find all of the available services on a
// single domain.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-9
func AbsoluteTypeEnumerationDomain(domain string) string {
	return domainname.Absolute(
		RelativeTypeEnumerationDomain(),
		domain,
	)
}

// RelativeTypeEnumerationDomain returns the DNS name that is queried
// to perform "service type enumeration" relative to the domain in which the
// records are advertised.
//
// Service type enumeration is used to find all of the available services on a
// single domain.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-9
func RelativeTypeEnumerationDomain() string {
	return domainname.Relative("_services", "_dns-sd", "_udp")
}

// AbsoluteInstanceEnumerationDomain returns the DNS name that is queried to
// perform "service instance enumeration" (aka "browsing") for specific service
// type & domain.
//
// Service instance enumeration is used to find all of the instances of a
// specific service type on a specific domain.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-4.
func AbsoluteInstanceEnumerationDomain(service, domain string) string {
	return domainname.Absolute(
		RelativeInstanceEnumerationDomain(service),
		domain,
	)
}

// RelativeInstanceEnumerationDomain returns the DNS name that is queried to
// perform "service instance enumeration" (aka "browsing") for specific service
// type relative to the domain in which the records are advertised.
//
// Service instance enumeration is used to find all of the instances of a
// specific service type on a specific domain.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-4.
func RelativeInstanceEnumerationDomain(service string) string {
	return domainname.Relative(service)
}

// AbsoluteSelectiveInstanceEnumerationDomain returns the absolute DNS name that
// is queried to perform "selective instance enumeration" for a specific service
// sub-type.
//
// Selective instance enumeration is like instance enumeration (browsing), but
// the results are filtered even further to service instances where the service
// behaves in a specific way or fulfills some specific function.
//
// For example, browsing can be used to find all instances that provide the
// _http._tcp service type, but selective instance enumeration can be used to
// narrow those results to include only web servers that are printer control
// panels.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-7.1
func AbsoluteSelectiveInstanceEnumerationDomain(subType, serviceType, domain string) string {
	return domainname.Absolute(
		RelativeSelectiveInstanceEnumerationDomain(subType, serviceType),
		domain,
	)
}

// RelativeSelectiveInstanceEnumerationDomain returns the DNS name that is
// queried to perform "selective instance enumeration" for a specific service
// sub-type relative to the domain in which the records are advertised.
//
// Selective instance enumeration is like instance enumeration (browsing), but
// the results are filtered even further to service instances where the service
// behaves in a specific way or fulfills some specific function.
//
// For example, browsing can be used to find all instances that provide the
// _http._tcp service type, but selective instance enumeration can be used to
// narrow those results to include only web servers that are printer control
// panels.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-7.1
func RelativeSelectiveInstanceEnumerationDomain(subType, serviceType string) string {
	return domainname.Relative(
		subType,
		"_sub",
		RelativeInstanceEnumerationDomain(serviceType),
	)
}
