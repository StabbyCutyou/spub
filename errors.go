package spub

import "fmt"

// 6 CustomErrors
// 3 Behavioral Interfaces (2 distinct, 1 composed of both)

// NoSubscriberID is returned when there is no subscriber id
// like during a specific publish / broadcast error
const NoSubscriberID = "-1"

// HasSubscriber represents an error with a SubscriberID
type HasSubscriber interface {
	ID() string
}

// HasMessage represents an error with a Data payload
type HasMessage interface {
	Message() []byte
}

// HasSubscriberAndMessage represents an error with both a SubscriberID and a
// Data payload
type HasSubscriberAndMessage interface {
	HasSubscriber
	HasMessage
}

// ErrPublishDeadline is returned when the message cannot be published due to
// timeout
type ErrPublishDeadline struct {
	Err          error
	Data         []byte
	SubscriberID string
}

// ErrShuttingDown is returned when the message cannot be published due to
// shutdown
type ErrShuttingDown struct {
	Data          []byte
	SubscriberID  string
	FullBroadcast bool
}

// ErrDuplicateSubscriberID is returned when you register conflicting
// Subscriber ids
type ErrDuplicateSubscriberID struct {
	SubscriberID string
}

// ErrSubscriberWithoutID is returned when you register a Subscriber
// without an ID
type ErrSubscriberWithoutID struct {
	SubscriberID string
}

// ErrUnknownSubscriber is returned when the message cannot be published due
// to a Subscriber being unknown by ID
type ErrUnknownSubscriber struct {
	Data         []byte
	SubscriberID string
}

// ID returns the Subscriber id
func (e ErrPublishDeadline) ID() string {
	return e.SubscriberID
}

// Message returns a data payload
func (e ErrPublishDeadline) Message() []byte {
	return e.Data
}

func (e ErrPublishDeadline) Error() string {
	return fmt.Sprintf("error publishing data, deadline hit: %s", e.Err)
}

// ID returns the
func (e ErrShuttingDown) ID() string {
	if e.FullBroadcast {
		return "-1"
	}
	return e.SubscriberID
}

// Message returns a data payload
func (e ErrShuttingDown) Message() []byte {
	return e.Data
}

func (e ErrShuttingDown) Error() string {
	return "error publishing data, shutting down"
}

// ID returns the Subscriber id
func (e ErrDuplicateSubscriberID) ID() string {
	return e.SubscriberID
}

func (e ErrDuplicateSubscriberID) Error() string {
	return fmt.Sprintf("error registering Subscriber, duplicate Subscriber ID: %s", e.SubscriberID)
}

// ID returns the Subscriber id
func (e ErrSubscriberWithoutID) ID() string {
	return e.SubscriberID
}

func (e ErrSubscriberWithoutID) Error() string {
	return "error registering Subscriber, you must provide a unique ID when registering a Subscriber"
}

// ID returns the Subscriber id
func (e ErrUnknownSubscriber) ID() string {
	return e.SubscriberID
}

// Message returns a data payload
func (e ErrUnknownSubscriber) Message() []byte {
	return e.Data
}

func (e ErrUnknownSubscriber) Error() string {
	return fmt.Sprintf("error registering Subscriber, unknown Subscriber ID: %s", e.SubscriberID)
}
