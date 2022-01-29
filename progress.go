package main

import (
	"fmt"
	"strings"
	"sync"
)

var (
	chPlanned = make(chan struct{})
	chDone    = make(chan struct{})
)

func progressMeter(wg *sync.WaitGroup, stop <-chan struct{}) {
	defer wg.Done()
	planned, done := 0, 0
loop:
	for {
		select {
		case <-stop:
			break loop
		case <-chPlanned:
			planned++
		case <-chDone:
			done++
		}
		fmt.Printf("\rLoaded %d of %d%s", done, planned, strings.Repeat(" ", 8))
	}
	fmt.Printf("\rAll %d files has been processed\n", done)
}
