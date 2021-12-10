package mdns

import (
	"net"

	"golang.org/x/net/ipv6"
)

var (
	// ipv6Group is the multicast group used for mDNS over IPv6.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	ipv6Group = net.ParseIP("ff02::fb")

	// ipv6GroupAddress is the address to which mDNS queries are sent when using
	// IPv6.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	ipv6GroupAddress = &net.UDPAddr{IP: ipv6Group, Port: Port}

	// ipv6ListenAddress is the address to which the mDNS server binds when
	// using IPv6.
	//
	// Note that the multicast group address is NOT used in order to control
	// more precisely which network interfaces join the multicast group.
	ipv6ListenAddress = &net.UDPAddr{
		IP:   net.ParseIP("ff02::"),
		Port: Port,
	}
)

// ipv6Transport is an IPv6-based UDP transport.
type ipv6Transport struct {
	pc *ipv6.PacketConn
}

// Listen starts listening for UDP packets on the given interfaces.
func (t *ipv6Transport) Listen(iface *net.Interface) (err error) {
	addr := ipv6ListenAddress
	conn, err := net.ListenUDP("udp6", addr)
	if err != nil {
		return err
	}

	t.pc = ipv6.NewPacketConn(conn)
	defer func() {
		if err != nil {
			t.pc.Close()
		}
	}()

	if err := t.pc.SetControlMessage(ipv6.FlagInterface, true); err != nil {
		return err
	}

	return t.pc.JoinGroup(
		iface,
		&net.UDPAddr{
			IP: ipv6Group,
		},
	)
}

// // Read reads the next packet from the transport.
// func (t *ipv6Transport) Read() (*InboundPacket, error) {
// 	buf := getBuffer()

// 	n, cm, src, err := t.pc.ReadFrom(buf)
// 	if err != nil {
// 		putBuffer(buf)
// 		logReadError(t.Logger, t.Group(), err)
// 		return nil, err
// 	}

// 	if cm == nil {
// 		putBuffer(buf)
// 		err := fmt.Errorf("empty control message from %s", src)
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
// func (t *ipv6Transport) Write(p *OutboundPacket) error {
// 	if _, err := t.pc.WriteTo(
// 		p.Data,
// 		&ipv6.ControlMessage{
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
func (t *ipv6Transport) Group() *net.UDPAddr {
	return ipv6GroupAddress
}

// Close closes the transport, preventing further reads and writes.
func (t *ipv6Transport) Close() error {
	return t.pc.Close()
}
