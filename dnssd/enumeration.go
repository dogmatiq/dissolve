package dnssd

// TypeEnumDomain returns the DNS name that is queried to perform
// "service type enumeration" for a specific domain.
//
// Service type enumeration is used to find all of the available services on a
// single domain.
//
// See https://datatracker.ietf.org/doc/html/rfc6763#section-9
func TypeEnumDomain(domain string) string {
	return "_services._dns-sd._udp." + domain
}

// InstanceEnumDomain returns the DNS name that is queried to perform "service
// instance enumeration" (aka "browsing") for specific service & domain.
//
// Service instance enumeration is used to find all of the instances of a
// specific service on a specific domain.
//
// See https://tools.ietf.org/html/rfc6763#section-4.
func InstanceEnumDomain(service, domain string) string {
	return service + "." + domain
}

// SelectiveInstanceEnumDomain returns the DNS name that is queried to perform
// "selective instance enumeration" for a specific service sub-type.
//
// Selective instance enumeration is like instance enumeration (browsing), but
// the results are filtered even further to service instances where the service
// behaves in a specific way or fulfills some specific function.
//
// For example, browsing can be used to find all instances that provide the
// _http._tcp service, but selective instance enumeration can be used to narrow
// those results to include only web servers that are printer control panels.
//
// See https://tools.ietf.org/html/rfc6763#section-7.1
func SelectiveInstanceEnumDomain(subtype, service, domain string) string {
	return subtype + "._sub." + InstanceEnumDomain(service, domain)
}
