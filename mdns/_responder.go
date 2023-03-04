package mdns

import (
	"context"
	"sync"

	"github.com/miekg/dns"
)

// Port is the multicast DNS UDP port number.
const Port = 5353

// type Observer interface {
// 	UniqueNameAcquired(name string) []dns.RR
// 	UniqueNameReleased(name string)
// }

type Responder struct {
	m                sync.Mutex
	count            int
	announce, cancel chan *announceRequest
	done             chan struct{}
}

type announceRequest struct {
	records []dns.RR
	result  chan error
}

func (r *Responder) AnnounceShared(ctx context.Context, records []dns.RR) error {
	r.start()
	defer r.stop()

	req := &announceRequest{
		records: records,
		result:  make(chan error, 1),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case r.announce <- req:
	}

	select {
	case <-ctx.Done():
		r.cancel <- req
		return ctx.Err()
	case err := <-req.result:
		return err
	}
}

func (r *Responder) start() {
	r.m.Lock()
	defer r.m.Unlock()

	r.count++

	if r.count != 1 {
		return
	}

	r.announce = make(chan *announceRequest)
	r.cancel = make(chan *announceRequest)
	r.done = make(chan struct{})

	go r.run()
}

func (r *Responder) stop() {
	r.m.Lock()
	defer r.m.Unlock()

	r.count--

	if r.count != 0 {
		return
	}

	close(r.announce)
	<-r.done
}

func (r *Responder) run() {
	defer close(r.done)

	for {
		select {
		case req, ok := <-r.announce:
			if !ok {
				return
			}

		}
	}
}

// func (r *Responder) AnnounceUnique(
// 	ctx context.Context,
// 	ready func() []dns.RR,
// ) error {
// }

// func (r *Responder) AcquireUniqueName(name string) {
// 	r.init()

// 	select {
// 	case r.commands <- func(r *Responder) error { return r.acquire(name) }:
// 	case <-r.done:
// 	}
// }

// func (r *Responder) ReleaseUniqueName(name string) {
// 	r.init()

// 	select {
// 	case r.commands <- func(r *Responder) error { return r.release(name) }:
// 	case <-r.done:
// 	}
// }

// // Run responds to multicast DNS queries until ctx is canceled or an error
// // occurs.
// func (r *Responder) Run(ctx context.Context) error {
// 	r.init()
// 	defer close(r.done)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		case cmd := <-r.commands:
// 			if err := cmd(r); err != nil {
// 				return err
// 			}
// 		}
// 	}
// }

// func (r *Responder) init() {
// 	r.once.Do(func() {
// 		r.commands = make(chan func(r *Responder) error)
// 		r.done = make(chan struct{})
// 	})
// }

// func (r *Responder) acquire(name string) error {
// 	return nil
// }

// func (r *Responder) release(name string) error {
// 	return nil
// }
