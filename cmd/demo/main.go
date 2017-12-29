package main

import (
	"sync"
	"time"

	"github.com/StabbyCutyou/spub"
)

func main() {
	p := spub.New(time.Millisecond * 1)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for range p.Err() {
		}
	}()
	for i := 0; i < 2; i++ {
		l := spub.Listener{
			C: make(chan []byte),
		}
		p.Register(l)
		go func(il *spub.Listener) {
			for range il.C {
			}
			wg.Done()
		}(&l) // pop off endlessly
	}
	for i := 0; i < 100000; i++ {
		p.Send(make([]byte, 0))
	}
	p.Close()
	wg.Wait()
}
