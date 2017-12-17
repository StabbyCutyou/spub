# Spub
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/StabbyCutyou/spub)

# What is Spub?
Spub is a simple, channel-based asynchronous broadcast publishing wrapper.

With Spub, you can send once and have any number of receivers get a copy of the data.

# How do I use it?

First, create a spub.Mux by passing it a time.Duration to act as the default timeout to use to determine when a send call should be considered invalid:

```go
m := spub.New(time.Milisecond * 1)
```

You will want to begin receiving on the errors channel in a go-routine so you can handle any timeouts or other delivery failures on your own. Spub will return several typed errors that contain the data that wasn't able to be sent, and if applicable, data identifying the Listener that was unable to be sent to, so you can manage retries on your own.

Note that this errors channel will never be closed, because it is used to report issues during shutdown as well as normal operations.

Take care not to assume this channel will ever close

```go
go func() {
    for err := <-m.Err() {
        switch err.(type) {
            case ErrPublishDeadline:
            case ErrShuttingDown:
            case ErrUnknownListener:
        }
    }
}
```

You can then begin registering Listeners. Listeners must have an ID, and that ID must be unique amongst all other Listeners. This is to support the case where 1 out o N Listeners fails to receive a message in time, or that the Mux is shutting down partway through a send, and you need to know which Listeners did not get the messages.

```go
l := spub.Listener{
    ID: "unique-id", // not optional
    C: make(chan []byte), // optional, will default to this value
    Timeout: time.Milisecond * 1 // optional, will default to the Mux timeout you specified
}

err := m.Register(l)
if err != nil {
    // You probably didn't set the ID or you set a duplicate ID
}
```

Now, you can begin to Send to the Listeners. Note that any erros from Send or SendTo will always be reported via the Err() channel, never synchronously. Both Send and SendTo are meant to behave in an asynchronous, non-blocking manner.

```go
m.Send([]byte("some really important data"))
```

If you receive any errors in your Error handler, you can use this method to attempt to retry:

```go
// err is either ErrPublishDeadline, ErrShuttingDown, ErrUnknownListener
m.SendTo(err.Data, err.ListenerID)
```