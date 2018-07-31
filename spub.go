// Package spub has some pub sub stuff, so please enjoy it
package spub

import (
	"context"
	"sync"
	"time"
)

// Subscriber wraps a channel to receive bytes from
type Subscriber struct {
	ID      string
	C       chan []byte
	Timeout time.Duration
	quit    chan struct{}
}

// Publisher registers a Subscriber and produces errors
type Publisher struct {
	// Nice little mutex 👒
	mu          sync.Mutex
	subscribers map[string]Subscriber

	err            chan error
	defaultTimeout time.Duration
	quit           chan struct{}
}

// New returns a new Publisher
func New(d time.Duration) *Publisher {
	return &Publisher{
		subscribers:    make(map[string]Subscriber),
		err:            make(chan error),
		defaultTimeout: d,
		quit:           make(chan struct{}),
	}
}

// Subscribe will register a subscriber
func (p *Publisher) Subscribe(s Subscriber) error {
	if s.ID == "" {
		return ErrSubscriberWithoutID{}
	}
	if s.C == nil {
		s.C = make(chan []byte)
	}
	if s.Timeout == 0 {
		s.Timeout = p.defaultTimeout
	}
	s.quit = make(chan struct{})
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.subscribers[s.ID]; ok {
		return ErrDuplicateSubscriberID{SubscriberID: s.ID}
	}
	p.subscribers[s.ID] = s
	return nil
}

// SendTo allows you to send to just 1 specific subscriber, good for retrying failures
func (p *Publisher) SendTo(b []byte, id string) {
	if p.Stopped() {
		go func() { p.err <- ErrShuttingDown{Data: b, SubscriberID: id} }()
	}
	p.mu.Lock()
	l, ok := p.subscribers[id]
	p.mu.Unlock()
	if !ok {
		p.err <- ErrUnknownSubscriber{Data: b, SubscriberID: l.ID}
		return
	}
	go p.sendto(b, &l)
}

// Broadcast will produce a message to the subscribers. Check Err() for errors
func (p *Publisher) Broadcast(b []byte) {
	if p.Stopped() {
		go func() { p.err <- ErrShuttingDown{Data: b, FullBroadcast: true} }()
	}
	// For the range of subscribers
	p.mu.Lock()
	for _, l := range p.subscribers {
		// Fire off a goroutine
		// Need to make a local copy, or else we run into closure conundrums
		lx := l
		go p.sendto(b, &lx)
	}
	p.mu.Unlock()
}

func (p *Publisher) sendto(b []byte, s *Subscriber) {
	ctx, cncl := context.WithTimeout(context.Background(), s.Timeout)
	defer cncl()
	defer func() {
		// TODO find a way around the panic on closed send
		// Can't find a way to signal to callers the subscriber is closed
		// so they can break out of a loop without compromising the fact
		// multiple go-routines might be trying to write to it
		// other than removing concurrency from send, but i'd prefer not
		// Since this is an edge case on shutdown only, and we can give them
		// the data back, this is ok for now
		if err := recover(); err != nil {
			p.err <- ErrShuttingDown{Data: b, SubscriberID: s.ID}
		}
	}()
	// Now send on the subscriber, hit the deadline, or bail on a closed Publisher
	select {
	case <-p.quit:
		p.err <- ErrShuttingDown{Data: b, SubscriberID: s.ID}
	case <-ctx.Done(): // If timeout, error
		p.err <- ErrPublishDeadline{Err: ctx.Err(), Data: b, SubscriberID: s.ID}
	case s.C <- b: // if it sent, we're done
	}
}

// Stop closes the Publisher after stopping the subscribers
func (p *Publisher) Stop() {
	close(p.quit)
	p.mu.Lock()
	for _, l := range p.subscribers {
		// If we try to get clever and close inside of Sends select on quit
		// we can run into cases where the channel never gets closed due to the
		// timing of the scheduler
		close(l.quit)
		close(l.C)
	}
	p.mu.Unlock()
}

// Stopped returns whether or not the publisher has been stopped
func (p *Publisher) Stopped() bool {
	select {
	case <-p.quit:
		return true
	default:
		return false
	}
}

// Unsubscribe will turn a subscriber off and close it's channels
func (p *Publisher) Unsubscribe(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	l, ok := p.subscribers[id]
	if !ok {
		return ErrUnknownSubscriber{SubscriberID: id}
	}
	close(l.quit)
	close(l.C)
	delete(p.subscribers, id)
	return nil
}

// Err returns a channel that emits errors
// This channel never closes
func (p *Publisher) Err() <-chan error {
	return p.err
}
