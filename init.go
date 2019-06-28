package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/fatih/pool"
	"github.com/sandman-cs/core"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// Configuration File Opjects
type configuration struct {
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

var (
	conn             *amqp.Connection
	rabbitCloseError chan *amqp.Error
	conf             configuration
	messages         = make(chan string, 128)
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

	//Load Configuration Data
	dat, _ := ioutil.ReadFile("conf.json")
	err := json.Unmarshal(dat, &conf)
	core.CheckError(err)

	for i := 0; i < 5; i++ {
		go func() {
			for {
				sendUDPMessage()
				time.Sleep(100 * time.Millisecond)
			}
		}()
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

func sendUDPMessage() {

	serverAddr, err := net.ResolveUDPAddr("udp", conf.DstSrv+":"+conf.DstPort)
	core.CheckError(err)
	localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	core.CheckError(err)
	conn, err := net.DialUDP("udp", localAddr, serverAddr)
	core.CheckError(err)

	if err == nil {

		defer conn.Close()
		for {
			msg := <-messages
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
