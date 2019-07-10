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
	conn             *amqp.Connection
	rabbitCloseError chan *amqp.Error
	conf             configuration
	messages         = make(chan string, 128)
	messageArr       []chan string
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

	//Load Configuration Data
	dat, _ := ioutil.ReadFile("conf.json")
	err := json.Unmarshal(dat, &conf)
	CheckError(err)

	if len(conf.Channels) > 0 {
		fmt.Println("Launching with new configuration")
		for i := 0; i < len(conf.Channels); i++ {
			messageArr[i] = make(chan string, 128)

		}
		os.Exit(0)
	} else {

		//Legacy Launch
		for i := 0; i < conf.ChannelCount; i++ {
			go func() {
				for {
					sendUDPMessage(conf.DstSrv, conf.DstPort, messages)
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
		msg := <-messages
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
