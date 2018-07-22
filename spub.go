// Package spub has some pub sub stuff
package spub

import (
	"context"
	"sync"
	"time"
)

// Listener is a channel to receive bytes from and a context for cancellation
type Listener struct {
	ID      string
	C       chan []byte
	Timeout time.Duration
	quit    chan struct{}
}

// Pub registers listeners and produces errors
type Pub struct {
	mu        sync.Mutex
	listeners map[string]Listener

	err            chan error
	defaultTimeout time.Duration
	quit           chan struct{}
}

// New returns a new Pub
func New(d time.Duration) *Pub {
	return &Pub{
		listeners:      make(map[string]Listener),
		err:            make(chan error),
		defaultTimeout: d,
		quit:           make(chan struct{}),
	}
}

// Subscribe will register a listener
func (p *Pub) Subscribe(l Listener) error {
	if l.ID == "" {
		return ErrListenerWithoutID{}
	}
	if l.C == nil {
		l.C = make(chan []byte)
	}
	if l.Timeout == 0 {
		l.Timeout = p.defaultTimeout
	}
	l.quit = make(chan struct{})
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.listeners[l.ID]; ok {
		return ErrDuplicateListenerID{ListenerID: l.ID}
	}
	p.listeners[l.ID] = l
	return nil
}

// SendTo allows you to send to just 1 specific listener, good for retrying failures
func (p *Pub) SendTo(b []byte, id string) {
	if p.Stopped() {
		go func() { p.err <- ErrShuttingDown{Data: b, ListenerID: id} }()
	}
	p.mu.Lock()
	l, ok := p.listeners[id]
	p.mu.Unlock()
	if !ok {
		p.err <- ErrUnknownListener{Data: b, ListenerID: l.ID}
		return
	}
	go p.sendto(b, &l)
}

// Broadcast will produce a message to the listeners. Check Err() for errors
func (p *Pub) Broadcast(b []byte) {
	if p.Stopped() {
		go func() { p.err <- ErrShuttingDown{Data: b, FullBroadcast: true} }()
	}
	// For the range of listeners
	p.mu.Lock()
	for _, l := range p.listeners {
		// Fire off a goroutine
		lx := l // Need to make a local copy, or else we run into closure conundrums
		go p.sendto(b, &lx)
	}
	p.mu.Unlock()
}

func (p *Pub) sendto(b []byte, l *Listener) {
	ctx, cncl := context.WithTimeout(context.Background(), l.Timeout)
	defer cncl()
	defer func() {
		// TODO find a way around the panic on closed send
		// Can't find a way to signal to callers the listener is closed
		// so they can break out of a loop without compromising the fact
		// multiple go-routines might be trying to write to it
		// other than removing concurrency from send, but i'd prefer not
		// Since this is an edge case on shutdown only, and we can give them
		// the data back, this is ok for now
		if err := recover(); err != nil {
			p.err <- ErrShuttingDown{Data: b, ListenerID: l.ID}
		}
	}()
	// Now send on the listener, hit the deadline, or bail on a closed mux
	select {
	case <-p.quit:
		p.err <- ErrShuttingDown{Data: b, ListenerID: l.ID}
	case <-ctx.Done(): // If timeout, error
		p.err <- ErrPublishDeadline{Err: ctx.Err(), Data: b, ListenerID: l.ID}
	case l.C <- b: // if it sent, we're done
	}
}

// Stop closes the mux after stopping the listeners
func (p *Pub) Stop() {
	close(p.quit)
	p.mu.Lock()
	for _, l := range p.listeners {
		// If we try to get clever and close inside of Sends select on quit
		// we can run into cases where the channel never gets closed due to the
		// timing of the scheduler
		close(l.quit)
		close(l.C)
	}
	p.mu.Unlock()
}

// Stopped returns whether or not the publisher has been stopped
func (p *Pub) Stopped() bool {
	select {
	case <-p.quit:
		return true
	default:
		return false
	}
}

// Unsubscribe will turn a listener off and close it's channels
func (p *Pub) Unsubscribe(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	l, ok := p.listeners[id]
	if !ok {
		return ErrUnknownListener{ListenerID: id}
	}
	close(l.quit)
	close(l.C)
	delete(p.listeners, id)
	return nil
}

// Err returns a channel that emits errors
// This channel never closes
func (p *Pub) Err() <-chan error {
	return p.err
}
