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

func processPayload(payload []byte) error {
	messages[0] <- string(payload[:])
	return nil
}
