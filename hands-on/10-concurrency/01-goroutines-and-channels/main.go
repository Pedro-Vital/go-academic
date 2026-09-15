package main

import (
	"fmt"
	"time"
)

func greet(phrase string) {
	fmt.Println("Hello!", phrase)
}

func slowGreet(phrase string, doneChan chan bool) {
	time.Sleep(3 * time.Second) // simulate a slow, long-taking task
	fmt.Println("Hello!", phrase)
	doneChan <- true // We could use any value, including false
}

func main() {
	go greet("Nice to meet you!")
	// Every function can be a goroutine
	go greet("How are you?")
	done := make(chan bool)
	go slowGreet("How ... are ... you ...?", done)
	go greet("I hope you're liking the course!")
	<-done
	// The program waits the value passed through the channel.
	// Precisely: main blocks until a value can be received from done.
	// The program ends only after it is returned
}