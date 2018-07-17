package main

import (
	"time"
	"sync"
	"fmt"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(10)
	go func() {
		for {
			fmt.Println("goroutine 1")
			wg.Done()
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			fmt.Println("goroutine2")
			wg.Done()
			time.Sleep(time.Second)
		}
	}()

	wg.Wait()
}
