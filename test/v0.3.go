package main

import (
	"fmt"
	"time"
)

func f(from string) {
	for i := 0; i < 3; i++ {
		fmt.Println(from, ":", i)
	}
}
func main() {
	// Create a new channel with `make(chan val-type)`.
	// Channels are typed by the values they convey.
	messages := make(chan string)

	tick := time.Tick(1 * time.Second)
	boom := time.After(3 * time.Second)

	// _Send_ a value into a channel using the `channel <-`
	// syntax. Here we send `"ping"`  to the `messages`
	// channel we made above, from a new goroutine.
	go func() {
		for {
			select {
			case <-tick:
				println("tick")
			case <-boom:
				messages <- "ping"
				return
			}
		}
	}() //使用"<-"向通道发送消息

	// The `<-channel` syntax _receives_ a value from the
	// channel. Here we'll receive the `"ping"` message
	// we sent above and print it out.
	msg := <-messages //"从通道读取数据"

	fmt.Println(msg)
	fmt.Scanln()
}
