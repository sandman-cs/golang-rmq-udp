package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

func main() {

	amqpURI := "amqp://" + conf.BrokerUser + ":" + conf.BrokerPwd + "@" + conf.Broker + conf.BrokerVhost

	// create the rabbitmq error channel
	rabbitCloseError = make(chan *amqp.Error)

	// run the callback in a separate thread
	go rabbitConnector(amqpURI)

	// establish the rabbitmq connection by sending
	// an error and thus calling the error callback

	rabbitCloseError <- amqp.ErrClosed

	for conn == nil {
		fmt.Println("Waiting to RabbitMQ Connection...")
		time.Sleep(5 * time.Second)
	}

	for i := 0; i <= conf.ChannelCount-1; i++ {
		tID := i // Passing I into a new variable for clean input to inline go func()
		go func() {
			threadID := tID // Passing back to variable name so it's static for loop below.
			for {
				OpenChannel(conn, threadID)
				log.Println("rabbit-listen closed with connection loss.")
			}
		}()
	}
	forever := make(chan bool)
	<-forever
	_ = conn.Close()
	os.Exit(0)
}

func processPayload(payload []byte) error {
	messages <- string(payload[:])
	return nil
}
