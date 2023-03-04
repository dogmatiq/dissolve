package mdns

import "net"

var (
	// IPv4Group is the multicast group used for mDNS over an IPv6 transport.
	//
	// See https://www.rfc-editor.org/rfc/rfc6762#section-3.
	IPv4Group = net.IPv4(224, 0, 0, 251)

	// IPv4GroupAddress is the address to which mDNS queries are over an IPv6
	// transport.
	//
	// See https://www.rfc-editor.org/rfc/rfc6762#section-3.
	IPv4GroupAddress = &net.UDPAddr{
		IP:   IPv4Group,
		Port: Port,
	}
)
