package spub

import "fmt"

// Returned when there is no listener id, like during a specific publish / broadcast error
const NoListenerID = "-1"

// HasListener represents an error with a ListenerID
type HasListener interface {
	ID() string
}

// HasMessage represents an error with a Data payload
type HasMessage interface {
	Message() []byte
}

// HasListenerAndMessage represents an error with both a ListenerID and a Data payload
type HasListenerAndMessage interface {
	HasListener
	HasMessage
}

// ErrPublishDeadline is returned when the message cannot be published due to timeout
type ErrPublishDeadline struct {
	Err        error
	Data       []byte
	ListenerID string
}

// ID returns the listener id
func (e ErrPublishDeadline) ID() string {
	return e.ListenerID
}

// Message returns a data payload
func (e ErrPublishDeadline) Message() []byte {
	return e.Data
}

func (e ErrPublishDeadline) Error() string {
	return fmt.Sprintf("error publishing data, deadline hit: %s", e.Err)
}

// ErrShuttingDown is returned when the message cannot be published due to shutdown
type ErrShuttingDown struct {
	Data          []byte
	ListenerID    string
	FullBroadcast bool
}

// ID returns the
func (e ErrShuttingDown) ID() string {
	if e.FullBroadcast {
		return "-1"
	}
	return e.ListenerID
}

// Message returns a data payload
func (e ErrShuttingDown) Message() []byte {
	return e.Data
}

func (e ErrShuttingDown) Error() string {
	return "error publishing data, shutting down"
}

// ErrDuplicateListenerID is returned when you register conflicting listener ids
type ErrDuplicateListenerID struct {
	ListenerID string
}

// ID returns the listener id
func (e ErrDuplicateListenerID) ID() string {
	return e.ListenerID
}

func (e ErrDuplicateListenerID) Error() string {
	return fmt.Sprintf("error registering listener, duplicate listener ID: %s", e.ListenerID)
}

// ErrListenerWithoutID is returned when you register a listener without an ID
type ErrListenerWithoutID struct {
	ListenerID string
}

// ID returns the listener id
func (e ErrListenerWithoutID) ID() string {
	return e.ListenerID
}

func (e ErrListenerWithoutID) Error() string {
	return "error registering listener, you must provide a unique ID when registering a listener"
}

// ErrUnknownListener is returned when the message cannot be published due to a listener being unknown by ID
type ErrUnknownListener struct {
	Data       []byte
	ListenerID string
}

// ID returns the listener id
func (e ErrUnknownListener) ID() string {
	return e.ListenerID
}

// Message returns a data payload
func (e ErrUnknownListener) Message() []byte {
	return e.Data
}

func (e ErrUnknownListener) Error() string {
	return fmt.Sprintf("error registering listener, unknown listener ID: %s", e.ListenerID)
}
