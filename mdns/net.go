package mdns

import "net"

// Port is the port on which mDNS queries are sent (and received, in the case of
// fully-compliant mDNS resolvers).
const Port = 5353

// multicastInterfaces returns a list of the system's network interfaces that
// support multicast.
func multicastInterfaces() ([]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	result := make([]net.Interface, 0, len(interfaces))
	for _, iface := range interfaces {
		if (iface.Flags & net.FlagMulticast) != 0 {
			result = append(result, iface)
		}
	}

	return result, nil
}
