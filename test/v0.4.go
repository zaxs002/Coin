package main

import (
	"time"
	"fmt"
)

func main() {
	c := make(chan string)

	go func() {
		time.Sleep(time.Second)
		c <- "one"
	}()
	go func() {
		time.Sleep(time.Second * 2)
		c <- "two"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg := <-c:
			fmt.Println("received", msg)
		//default:
		//	fmt.Println("...")
		}
	}
}
