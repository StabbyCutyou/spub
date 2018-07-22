package spub_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/StabbyCutyou/spub"
)

type benchCase struct {
	name      string
	listeners int
	data      []byte
}

var benchCases = []benchCase{
	{
		name:      "5 listeners",
		listeners: 5,
		data:      []byte(`aafkdjddfklldfldflkdflkdfkfdkdlkerieooeiereroieonbnvnn`),
	},
	{
		name:      "10 listeners",
		listeners: 10,
		data:      []byte(`aafkdjddfklldfldflkdflkdfkfdkdlkerieooeiereroieonbnvnn`),
	},
	{
		name:      "20 listeners",
		listeners: 20,
		data:      []byte(`aafkdjddfklldfldflkdflkdfkfdkdlkerieooeiereroieonbnvnn`),
	},
	{
		name:      "100 listeners",
		listeners: 100,
		data:      []byte(`aafkdjddfklldfldflkdflkdfkfdkdlkerieooeiereroieonbnvnn`),
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
					if errC%10000 == 0 {
						fmt.Printf("aaa------------------- %d ----------------------\n", errC)
					}
				}
			}()
			listeners := make([]spub.Listener, c.listeners)
			for i := 0; i > c.listeners; i++ {
				c := make(chan []byte)
				go func() {
					for range c {
						messC++
						if messC%10000 == 0 {
							fmt.Printf("-bbb------------------ %d ----------------------\n", messC)
						}
					}
				}() // pop off endlessly
				listeners[i].C = c
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
