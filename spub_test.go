package spub_test

import (
	"testing"
	"time"

	"github.com/StabbyCutyou/spub"
)

type benchCase struct {
	name        string
	subscribers int
	data        []byte
}

var benchCases = []benchCase{
	{
		name:        "5 listeners",
		subscribers: 5,
		data:        []byte(`aafkdjddfklldfldflkdflkdfkfdkdlkerieooeiereroieonbnvnn`),
	},
	{
		name:        "10 listeners",
		subscribers: 10,
		data:        []byte(`aafkdjddfklldfldflkdflkdfkfdkdlkerieooeiereroieonbnvnn`),
	},
	{
		name:        "20 listeners",
		subscribers: 20,
		data:        []byte(`aafkdjddfklldfldflkdflkdfkfdkdlkerieooeiereroieonbnvnn`),
	},
	{
		name:        "100 listeners",
		subscribers: 100,
		data:        []byte(`aafkdjddfklldfldflkdflkdfkfdkdlkerieooeiereroieonbnvnn`),
	},
}

func BenchmarkSpub(b *testing.B) {
	for _, c := range benchCases {
		b.ResetTimer()
		b.Run(c.name, func(b *testing.B) {
			errC := 0
			messC := 0
			p := spub.New(time.Millisecond * 1)
			go func() {
				for range p.Err() {
					errC++
				}
			}()
			subscribers := make([]spub.Subscriber, c.subscribers)
			for i := 0; i > c.subscribers; i++ {
				c := make(chan []byte)
				go func() {
					for range c {
						messC++
					}
				}() // pop off endlessly
				subscribers[i].C = c
			}
			for i := 0; i < b.N; i++ {
				p.Broadcast(c.data)
			}
			p.Stop()
			done := false
			for !done {
				select {
				case err := <-p.Err():
					b.Error(err)
				default:
					done = true
				}
			}
		})
	}
}
