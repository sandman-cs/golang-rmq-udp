package main

import (
	"fmt"
	"os"
)

func main() {

	forever := make(chan bool)
	fmt.Println("Running....")
	<-forever
	os.Exit(0)
}

func processPayload(payload []byte, index int) error {
	messages[index] <- string(payload[:])
	return nil
}
