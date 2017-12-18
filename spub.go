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

// Close closes the listener
func (l *Listener) Close() {
	close(l.quit)
}

// Closed returns whether or not the Listener is closed
func (l *Listener) Closed() bool {
	select {
	case <-l.quit:
		return true
	default:
		return false
	}
}

// Mux registers listeners and produces errors
type Mux struct {
	mu             sync.Mutex
	listeners      map[string]Listener
	err            chan error
	defaultTimeout time.Duration
	quit           chan struct{}
}

// New returns a new mux
func New(d time.Duration) *Mux {
	return &Mux{
		err:            make(chan error),
		defaultTimeout: d,
		quit:           make(chan struct{}),
	}
}

// Register will register a listener
func (m *Mux) Register(l Listener) error {
	if l.ID == "" {
		return ErrListenerWithoutID{}
	}
	if l.C == nil {
		l.C = make(chan []byte)
	}
	if l.Timeout == 0 {
		l.Timeout = m.defaultTimeout
	}
	l.quit = make(chan struct{})
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.listeners[l.ID]; ok {
		return ErrDuplicateListenerID{ListenerID: l.ID}
	}
	m.listeners[l.ID] = l
	return nil
}

// SendTo allows you to send to just 1 specific listener, good for retrying failures
func (m *Mux) SendTo(b []byte, id string) {
	m.mu.Lock()
	l, ok := m.listeners[id]
	m.mu.Unlock()
	go func() {
		if !ok {
			m.err <- ErrUnknownListener{Data: b, ListenerID: l.ID}
			return
		}
		m.sendto(b, &l)
	}()
}

// Send will produce a message to the listeners. Check Err() for errors
func (m *Mux) Send(b []byte) {
	// For the range of listeners
	m.mu.Lock()
	for _, l := range m.listeners {
		// Fire off a goroutine
		go m.sendto(b, &l)
	}
	m.mu.Unlock()
}

func (m *Mux) sendto(b []byte, l *Listener) {
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
			m.err <- ErrShuttingDown{Data: b, ListenerID: l.ID}
		}
	}()
	// Now send on the listener, hit the deadline, or bail on a closed mux
	select {
	case <-m.quit:
		m.err <- ErrShuttingDown{Data: b, ListenerID: l.ID}
	case <-ctx.Done(): // If timeout, error
		m.err <- ErrPublishDeadline{Err: ctx.Err(), Data: b, ListenerID: l.ID}
	case l.C <- b: // if it sent, we're done
	}
}

// Close closes the mux after draining the listeners?
func (m *Mux) Close() {
	close(m.quit)
	m.mu.Lock()
	for _, l := range m.listeners {
		// If we try to get clever and close inside of the Send select on quit
		// we can run into cases where the channel never gets closed due to the
		// timing of the scheduler
		l.Close()
		close(l.C)
	}
	m.mu.Unlock()
}

// RemoveListener is to turn a listener off and close it's channel
func (m *Mux) RemoveListener(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.listeners[id]
	if !ok {
		return ErrUnknownListener{ListenerID: id}
	}
	if !l.Closed() {
		l.Close()
	}
	delete(m.listeners, id)
	return nil
}

// Err returns a channel that emits errors
func (m *Mux) Err() chan error {
	return m.err
}
