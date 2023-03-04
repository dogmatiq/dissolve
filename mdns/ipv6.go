package mdns

import "net"

var (
	// IPv6Group is the multicast group used for mDNS over an IPv6 transport.
	//
	// See https://www.rfc-editor.org/rfc/rfc6762#section-3.
	IPv6Group = net.ParseIP("ff02::fb")

	// IPv6GroupAddress is the address to which mDNS queries are over an IPv6
	// transport.
	//
	// See https://www.rfc-editor.org/rfc/rfc6762#section-3.
	IPv6GroupAddress = &net.UDPAddr{IP: IPv6Group, Port: Port}
)
