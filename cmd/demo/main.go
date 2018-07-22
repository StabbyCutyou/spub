package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/StabbyCutyou/spub"
)

func main() {
	// Set a timeout that is reasonable as a default
	// Depending on how big of a mailbox you give your listeners,
	// and how busy your publishers are, this number becomes very
	// important to avoid bleeding messages through Err()
	p := spub.New(time.Minute * 1)
	// We'll use this to block on shutdown near the end
	wg := sync.WaitGroup{}
	// Dynamically control the number of listeners to demonstrate parallelism
	numListeners := 3
	wg.Add(numListeners)
	// Track why errors happen and for whom
	// -1 means it was not applicable to a listener directly
	errReasons := make(map[string]map[string]int)
	errReasons["-1"] = make(map[string]int)
	for i := 0; i < numListeners; i++ {
		// Seed the reason tracker
		errReasons[strconv.Itoa(i)] = make(map[string]int)
	}
	// Start draining Err now
	go func() {
		for err := range p.Err() {
			switch x := err.(type) {
			case spub.HasListener:
				m := errReasons[x.ID()]
				i, ok := m[err.Error()]
				if !ok {
					m[err.Error()] = 1
					continue
				}
				m[err.Error()] = i + 1
			}
		}
	}()
	for i := 0; i < numListeners; i++ {
		// Make a new listener
		l := spub.Listener{
			ID: strconv.Itoa(i),
			C:  make(chan []byte),
		}
		// Subscribe to the feed
		p.Subscribe(l)
		// Fire off a go routine to drain the listeners and keep count
		go drainListener(&l, &wg)
	}
	// Send off 100k messages. Recall that Broadcast spawns 1 go routine per listener
	// when sending messages
	for i := 0; i < 100000; i++ {
		p.Broadcast(make([]byte, 0))
	}
	// If you don't delay a bit, the final set of messages get caught in a shutdown filter
	// In the real world, you'll want to have a mechanism to continue to drain errors after shutdown
	// and then manage noting which messages never got sent where, for later replay or investigation
	time.Sleep(1 * time.Second)
	fmt.Println("about to stop")
	// Stop the publishing
	p.Stop()
	fmt.Println("stop called")
	// Wait for draining
	wg.Wait()
	for len(p.Err()) > 0 {
		// Let it drain - it will never close, but it will hit zero once stop is called
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("errors:\n%+v\n", errReasons)
}

func drainListener(il *spub.Listener, wg *sync.WaitGroup) {
	x := 0
	for range il.C {
		x++
	}
	fmt.Printf("done with %s, got %d msgs\n", il.ID, x)
	// This will happen once Stop is called
	wg.Done()
}
