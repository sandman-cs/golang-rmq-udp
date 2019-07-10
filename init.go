package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/fatih/pool"
	"github.com/streadway/amqp"
)

// Configuration File Opjects
type configuration struct {
	Channels     []sourceDest
	AppName      string
	AppVer       string
	ServerName   string
	Broker       string
	BrokerUser   string
	BrokerPwd    string
	BrokerQueue  string
	BrokerVhost  string
	LocalEcho    bool
	ChannelCount int
	ChannelSize  int
	DstSrv       string
	DstPort      string
}

type sourceDest struct {
	Broker       string
	BrokerUser   string
	BrokerPwd    string
	BrokerQueue  string
	BrokerVhost  string
	ChannelCount int
	DstSrv       string
	DstPort      string
}

var (
	conn             [255]*amqp.Connection
	rabbitCloseError []chan *amqp.Error
	conf             configuration
	messages         []chan string
	// Create a new instance of the logger. You can have any number of instances.
)

func init() {
	conf.AppName = "rabbit-listen-udp"
	conf.AppVer = "1.0"
	conf.ServerName, _ = os.Hostname()
	conf.Broker = "localhost"
	conf.BrokerVhost = "/"
	conf.LocalEcho = true
	conf.DstSrv = "172.24.38.181"
	conf.DstPort = "514"
	conf.ChannelCount = 1

	conf.ChannelSize = 128

	//Load Configuration Data
	dat, _ := ioutil.ReadFile("conf.json")
	err := json.Unmarshal(dat, &conf)
	CheckError(err)

	if len(conf.Channels) > 0 {
		//Keep this part, spawn all the cool new stuff...............................
		for index, element := range conf.Channels {
			//Load Defaults if needed
			if element.ChannelCount == 0 {
				element.ChannelCount = conf.ChannelCount
			}
			if len(element.DstPort) == 0 {
				element.DstPort = conf.DstPort
			}
			if len(element.DstSrv) == 0 {
				element.DstSrv = conf.DstSrv
			}
			// Create Channel and launch publish threads.......
			log.Println("Creating Channel #", index)
			messages = append(messages, make(chan string, conf.ChannelSize))
			//Spawn Sending threads for each configuration entry
			for i := 0; i < conf.ChannelCount; i++ {
				go func(element sourceDest, index int) {
					for {
						sendUDPMessage(element.DstSrv, element.DstPort, messages[index])
						time.Sleep(100 * time.Millisecond)
					}
				}(element, index)
			}
			go rmqRecThread(element.BrokerUser, element.BrokerPwd, element.Broker, element.BrokerVhost, element.BrokerQueue, index)
		}

	} else {

		//Legacy Launch
		messages = append(messages, make(chan string, conf.ChannelSize))
		for i := 0; i < conf.ChannelCount; i++ {
			go func() {
				for {
					sendUDPMessage(conf.DstSrv, conf.DstPort, messages[0])
					time.Sleep(100 * time.Millisecond)
				}
			}()
		}
	}
}

func sendTCPMessage() {

	// create a factory() to be used with channel based pool
	factory := func() (net.Conn, error) { return net.Dial("tcp", conf.DstSrv+":"+conf.DstPort) }
	p, err := pool.NewChannelPool(2, 5, factory)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

	for {
		msg := <-messages[0]
		//fmt.Println(msg)
		conn, err := p.Get()
		if err != nil {
			println("Dial failed:", err.Error())
			break
		}
		_, err = conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println(msg, err)
		}
		conn.Close()
		time.Sleep(5 * time.Millisecond)

	}
}

func sendUDPMessage(dest string, port string, input chan string) {

	serverAddr, err := net.ResolveUDPAddr("udp", dest+":"+port)
	CheckError(err)
	localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	CheckError(err)
	conn, err := net.DialUDP("udp", localAddr, serverAddr)
	CheckError(err)

	if err == nil {

		defer conn.Close()
		for {
			msg := <-input
			buf := []byte(msg + "\n")
			_, err := conn.Write(buf)
			if err != nil {
				log.Println(err)
				return
			}
			time.Sleep(time.Millisecond * 5)
		}
	}
}

func rmqRecThread(brokerUser string, brokerPwd string, brokerURL string, brokerVhost string, brokerQueue string, index int) {

	amqpURI := "amqp://" + brokerUser + ":" + brokerPwd + "@" + brokerURL + brokerVhost

	// create the rabbitmq error channel
	rabbitCloseError = append(rabbitCloseError, make(chan *amqp.Error))

	// run the callback in a separate thread
	go rabbitConnector(amqpURI, index)

	// establish the rabbitmq connection by sending
	// an error and thus calling the error callback

	rabbitCloseError[index] <- amqp.ErrClosed

	for conn[index] == nil {
		fmt.Println("Waiting to RabbitMQ Connection on index", index, "...")
		time.Sleep(5 * time.Second)
	}

	for i := 0; i <= conf.ChannelCount-1; i++ {
		tID := i // Passing I into a new variable for clean input to inline go func()
		go func() {
			threadID := tID // Passing back to variable name so it's static for loop below.
			for {
				OpenChannel(conn[index], brokerQueue, threadID)
				log.Println("rabbit-listen closed with connection loss.")
			}
		}()
	}
	forever := make(chan bool)
	<-forever
	_ = conn[index].Close()
}
