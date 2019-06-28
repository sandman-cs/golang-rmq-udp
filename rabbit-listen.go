package main

import (
	"time"

	"github.com/sandman-cs/core"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

//OpenChannel Opens communication Channel to Queue
func OpenChannel(conn *amqp.Connection, chanNumber int) {

	//Open channel to broker
	ch, err := conn.Channel()
	core.FailOnError(err, "Failed to open a channel")
	ch.Qos(10, 0, true)
	defer ch.Close()

	//Consume messages off the queue
	msgs, err := ch.Consume(
		conf.BrokerQueue, // queue
		conf.ServerName,  // consumer
		false,            // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	if err != nil {
		log.Error("Failed to register a consumer:", err)
		return
	}
	log.Info("Rabbit-listener instance started.")

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			err = processPayload(d.Body)
			if err != nil {
				core.SyslogCheckError(err)
				d.Ack(true)
			} else {
				d.Ack(false)
			}
		}
		core.SyslogSendJSON("Error, Lost AMQP Connection.")
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

		core.CheckError(err)
		core.SendMessage("Trying to reconnect to RabbitMQ")
		time.Sleep(500 * time.Millisecond)
	}
}

// re-establish the connection to RabbitMQ in case
// the connection has died
//
func rabbitConnector(uri string) {
	var rabbitErr *amqp.Error

	for {
		rabbitErr = <-rabbitCloseError
		if rabbitErr != nil {
			core.SendMessage("Connecting to RabbitMQ")
			conn = connectToRabbitMQ(uri)
			rabbitCloseError = make(chan *amqp.Error)
			conn.NotifyClose(rabbitCloseError)
		}
	}
}
