package mdns

import (
	"net"

	"golang.org/x/net/ipv4"
)

var (
	// ipv4Group is the multicast group used for mDNS over IPv4.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	ipv4Group = net.IPv4(224, 0, 0, 251)

	// ipv4GroupAddress is the address to which mDNS queries are sent when using
	// IPv4.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	ipv4GroupAddress = &net.UDPAddr{
		IP:   ipv4Group,
		Port: Port,
	}

	// ipv4ListenAddress is the address to which the mDNS server binds when
	// using IPv4.
	//
	// Note that the multicast group address is NOT used in order to control
	// more precisely which network interfaces join the multicast group.
	ipv4ListenAddress = &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 0),
		Port: Port,
	}
)

// ipv4Transport is an IPv4-based UDP transport.
type ipv4Transport struct {
	pc *ipv4.PacketConn
}

// Listen starts listening for UDP packets on the given interfaces.
func (t *ipv4Transport) Listen(iface *net.Interface) (err error) {
	conn, err := net.ListenUDP("udp4", ipv4ListenAddress)
	if err != nil {
		return err
	}

	t.pc = ipv4.NewPacketConn(conn)
	defer func() {
		if err != nil {
			t.pc.Close()
		}
	}()

	if err := t.pc.SetControlMessage(ipv4.FlagInterface, true); err != nil {
		return err
	}

	return t.pc.JoinGroup(
		iface,
		&net.UDPAddr{
			IP: ipv4Group,
		},
	)
}

// // Read reads the next packet from the transport.
// func (t *ipv4Transport) Read() (*InboundPacket, error) {
// 	buf := getBuffer()

// 	n, cm, src, err := t.pc.ReadFrom(buf)
// 	if err != nil {
// 		putBuffer(buf)
// 		logReadError(t.Logger, t.Group(), err)
// 		return nil, err
// 	}

// 	buf = buf[:n]

// 	return &InboundPacket{
// 		t,
// 		Endpoint{
// 			cm.IfIndex,
// 			src.(*net.UDPAddr),
// 		},
// 		buf,
// 	}, nil
// }

// // Write sends a packet via the transport.
// func (t *ipv4Transport) Write(p *OutboundPacket) error {
// 	if _, err := t.pc.WriteTo(
// 		p.Data,
// 		&ipvx.ControlMessage{
// 			IfIndex: p.Destination.InterfaceIndex,
// 		},
// 		p.Destination.Address,
// 	); err != nil {
// 		logWriteError(t.Logger, p.Destination.Address, t.Group(), err)
// 		return err
// 	}

// 	return nil
// }

// Group returns the multicast group address for this transport.
func (t *ipv4Transport) Group() *net.UDPAddr {
	return ipv4GroupAddress
}

// Close closes the transport, preventing further reads and writes.
func (t *ipv4Transport) Close() error {
	return t.pc.Close()
}
