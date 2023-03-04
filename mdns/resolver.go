package mdns

import "github.com/miekg/dns"

// Resolver is a client for making "continuous" multicast DNS queries.
//
// See https://www.rfc-editor.org/rfc/rfc6762#section-5.2.
type Resolver struct {
}

// NewSession returns a new session for making multicast DNS queries.
func (r *Resolver) NewSession(events chan<- Event) (*Session, error) {
	return &Session{}, nil
}

// Close ends all sessions and stops the resolver from processing any multicast
// DNS traffic.
func (r *Resolver) Close() error {
	return nil
}

// An Event describes a change to a multicast DNS record.
type Event interface{}

// RecordDiscovered is an event that indicates a record that is unknown to the
// subscriber has been discovered.
type RecordDiscovered struct {
	Record dns.RR
}

// RecordUpdated is an event that indicates a record known to the subscriber has
// changed.
type RecordUpdated struct {
	Record dns.RR
	Prior  dns.RR
}

// RecordGone is an event that indicates a record that is known to the
// subscriber has gone away.
type RecordGone struct {
	Record dns.RR
}

// Session is a client for making multicast DNS queries.
type Session struct {
}

// Subscribe enrolls the session to receive information about multicast DNS
// records for the given service name, class and record type.
func (s *Session) Subscribe(name string, class, types uint16) {
}

// Unsubscribe stops the session receiving information about multicast DNS
// records for the given service name, class and record type.
func (s *Session) Unsubscribe(name string, class, types uint16) {
}

// Close ends all subscriptions and closes the event channel.
func (s *Session) Close() error {
	return nil
}
