# Spub
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/StabbyCutyou/spub)

# What is Spub?
Spub is a simple, channel-based asynchronous broadcast publishing wrapper.

With Spub, you can send once and have any number of receivers get a copy of the data.

# How do I use it?

First, create a spub.Publisher by passing it a time.Duration to act as the default timeout to use to determine when a send call should be considered invalid:

```go
p := spub.New(time.Milisecond * 1)
```

You will want to begin receiving on the errors channel in a go-routine so you can handle any timeouts or other delivery failures on your own. Spub will return several typed errors that contain the data that wasn't able to be sent, and if applicable, data identifying the Subscriber that was unable to be sent to, so you can manage retries on your own.

Note that this errors channel will never be closed, because it is used to report issues during shutdown as well as normal operations.

Take care not to assume this channel will ever close

```go
go func() {
    for err := <-p.Err() {
        switch err.(type) {
            case ErrPublishDeadline:
            case ErrShuttingDown:
            case ErrUnknownSubscriber:
        }
    }
}
```

You can then begin registering Subscribers. Subscribers must have an ID, and that ID must be unique amongst all other Subscribers. This is to support the case where 1 out o N Subscribers fails to receive a message in time, or that the Publisher is shutting down partway through a send, and you need to know which Subscribers did not get the messages.

```go
s := spub.Subscriber{
    ID: "unique-id", // not optional
    C: make(chan []byte), // optional, will default to this value
    Timeout: time.Milisecond * 1 // optional, will default to the Publisher timeout you specified
}

if err := p.Subscribe(s); err != nil {
    // You probably didn't set the ID or you set a duplicate ID
}
```

Now, you can begin to Broadcast to the Subscribers. Note that any erros from Broadcast or SendTo will always be reported via the Err() channel, never synchronously. Both Broadcast and SendTo are meant to behave in an asynchronous, non-blocking manner.

```go
p.Broadcast([]byte("some really cool data"))
```

If you receive any errors in your Error handler, you can use this method to attempt to retry:

```go
// err is either ErrPublishDeadline, ErrShuttingDown, ErrUnknownSubscriber
p.SendTo(err.Data, err.SubscriberID)
```