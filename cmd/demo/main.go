package main

import (
	"sync"
	"time"

	"github.com/StabbyCutyou/spub"
)

func main() {
	m := spub.New(make(chan error), time.Millisecond*1)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for range m.Err() {
		}
	}()

	for i := 0; i < 2; i++ {
		l := spub.Listener{
			C: make(chan []byte),
		}
		m.Register(l)
		go func(il *spub.Listener) {

			for range il.C {
			}
			wg.Done()
		}(&l) // pop off endlessly
	}

	for i := 0; i < 100000; i++ {
		m.Send(make([]byte, 0))
	}
	m.Close()
	wg.Wait()
}
