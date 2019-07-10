package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

var (
	sysLogSrv  string
	sysLogPort string
	appName    string
	appVersion string
	serverName string
	localEcho  bool
)

func init() {

	sysLogSrv = "splunk"
	sysLogPort = "514"
	appName = "unknown"
	appVersion = "1.0"
	serverName, _ = os.Hostname()

}

//InitSysLog inititialize the stuff...
func InitSysLog(iSysLogSrv string, iSysLogPort string, iAppName string, iAppVersion string, iLocalEcho bool, iServerName string) {

	if iSysLogSrv != "" {
		sysLogSrv = iSysLogSrv
	}
	if iSysLogPort != "" {
		sysLogPort = iSysLogPort
	}
	if iAppName != "" {
		appName = iAppName
	}
	if iAppVersion != "" {
		appVersion = iAppVersion
	}
	if iServerName != "" {
		serverName = iServerName
	}
	if iLocalEcho {
		localEcho = true
	}

}

//SyslogSend - Send message to Syslog Dest
func SyslogSend(msg string) {

	if localEcho {
		log.Println(msg)
	}

	ServerAddr, err := net.ResolveUDPAddr("udp", sysLogSrv+":"+sysLogPort)
	CheckError(err)
	if err == nil {

		LocalAddr, err := net.ResolveUDPAddr("udp", ":0")
		CheckError(err)

		Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
		CheckError(err)

		defer Conn.Close()
		buf := []byte(msg)
		if _, err := Conn.Write(buf); err != nil {
			CheckError(err)
		}
	}
}

//SyslogSendJSON -- Send Json to syslog dest
func SyslogSendJSON(msg string) {

	payload := fmt.Sprintf("{\"app_name\":\"%s\",\"server\":\"%s\",\"msg\":\"%s\"}", appName, serverName, msg)
	SyslogSend(payload)
}

// SyslogCheckError -- Send error
func SyslogCheckError(pError error) {

	if pError != nil {
		payload := fmt.Sprintf("{\"app_name\":\"%s\",\"server\":\"%s\",\"error\":\"%s\"}", appName, serverName, pError)
		SyslogSend(payload)
	}
}

//SyslogCheckFuncError Adds additional info to Error Field
func SyslogCheckFuncError(pError error, sFunc string) {

	if pError != nil {

		payload := fmt.Sprintf("{\"app_name\":\"%s\",\"function\":\"%s\",\"server\":\"%s\",\"error\":\"%s\"}", appName, sFunc, serverName, pError)
		SyslogSend(payload)
	}
}

//CheckError function
func CheckError(err error) {
	if err != nil {
		log.Println("Error: ", err)
	}
}

//SendMessage to udp listener
func SendMessage(msg string) {
	if localEcho {
		fmt.Println(msg)

	}
}

//FailOnError - fail on error
func FailOnError(err error, msg string) {
	if err != nil {
		log.Println("Fatal Error", msg, err)
		SyslogSend(fmt.Sprintf("%s: %s", msg, err))
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
