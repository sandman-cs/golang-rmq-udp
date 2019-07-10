package main

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

//OpenChannel Opens communication Channel to Queue
func OpenChannel(conn *amqp.Connection, sQueue string, chanNumber int) {

	//Open channel to broker
	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	ch.Qos(10, 0, true)
	defer ch.Close()

	//Consume messages off the queue
	msgs, err := ch.Consume(
		sQueue,          // queue
		conf.ServerName, // consumer
		false,           // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Println("Failed to register a consumer:", err)
		return
	}
	log.Println("Rabbit-listener instance started.")

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			err = processPayload(d.Body)
			if err != nil {
				SyslogCheckError(err)
				d.Ack(true)
			} else {
				d.Ack(false)
			}
		}
		SyslogSendJSON("Error, Lost AMQP Connection.")
		forever <- false
	}()

	<-forever
}

// Try to connect to the RabbitMQ server as
// long as it takes to establish a connection
//
func connectToRabbitMQ(uri string) *amqp.Connection {
	for {
		conn, err := amqp.Dial(uri)

		if err == nil {
			return conn
		}

		CheckError(err)
		SendMessage("Trying to reconnect to RabbitMQ")
		time.Sleep(500 * time.Millisecond)
	}
}

// re-establish the connection to RabbitMQ in case
// the connection has died
//
func rabbitConnector(uri string, index int) {
	var rabbitErr *amqp.Error

	for {
		rabbitErr = <-rabbitCloseError[index]
		if rabbitErr != nil {
			SendMessage("Connecting to RabbitMQ")
			conn[index] = connectToRabbitMQ(uri)
			rabbitCloseError[index] = make(chan *amqp.Error)
			conn[index].NotifyClose(rabbitCloseError[index])
		}
	}
}
