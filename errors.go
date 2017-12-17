package spub

import "fmt"

// ErrPublishDeadline is returned when the message cannot be published due to timeout
type ErrPublishDeadline struct {
	Err        error
	Data       []byte
	ListenerID string
}

func (e ErrPublishDeadline) Error() string {
	return fmt.Sprintf("error publishing data, deadline hit: %s", e.Err)
}

// ErrShuttingDown is returned when the message cannot be published due to shutdown
type ErrShuttingDown struct {
	Data       []byte
	ListenerID string
}

func (e ErrShuttingDown) Error() string {
	return "error publishing data, shutting down"
}

// ErrDuplicateListenerID is returned when you register conflicting listener ids
type ErrDuplicateListenerID struct {
	ListenerID string
}

func (e ErrDuplicateListenerID) Error() string {
	return fmt.Sprintf("error registering listener, duplicate listener ID: %s", e.ListenerID)
}

// ErrListenerWithoutID is returned when you register a listener without an ID
type ErrListenerWithoutID struct {
	ListenerID string
}

func (e ErrListenerWithoutID) Error() string {
	return "error registering listener, you must provide a unique ID when registering a listener"
}

// ErrUnknownListener is returned when the message cannot be published due to a listener being unknown by ID
type ErrUnknownListener struct {
	Data       []byte
	ListenerID string
}

func (e ErrUnknownListener) Error() string {
	return fmt.Sprintf("error registering listener, unknown listener ID: %s", e.ListenerID)
}
