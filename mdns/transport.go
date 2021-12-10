package mdns

import "net"

// transport is an interface for communicating via UDP.
type transport interface {
	// Listen starts listening for UDP packets on the given interface.
	Listen(iface *net.Interface) error

	// // Read reads the next packet from the transport.
	// Read() (*InboundPacket, error)

	// // Write sends a packet via the transport.
	// Write(*OutboundPacket) error

	// Group returns the multicast group address for this transport.
	Group() *net.UDPAddr

	// Close closes the transport, preventing further reads and writes.
	Close() error
}
