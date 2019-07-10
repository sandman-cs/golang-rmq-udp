package main

import (
	"os"
)

func main() {

	forever := make(chan bool)
	<-forever
	_ = conn.Close()
	os.Exit(0)
}

func processPayload(payload []byte) error {
	messages[0] <- string(payload[:])
	return nil
}
