package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/nrm21/EtcdChat/support"
)

var version string // to be auto-added with -ldflags at build time

// Program entry point
func main() {
	println("Running Version: " + version)

	support.SetupCloseHandler() // setup ctrl + c to break loop
	println("Press ctrl + c to exit...")

	strIP := support.GetOutboundIP().String()
	config := getConfigContents("support/config.yml")
	clientID := generateID()
	//message := TakeUserInput()
	message := "this message was generated from " + runtime.GOOS
	timestamp := getMicroTime()
	keyToWrite := fmt.Sprintf("%s/%s", config.Etcd.BaseKeyToWrite, clientID)
	valueToWrite := fmt.Sprintf("%s | %s | %s", timestamp, strIP, message)

	// if localhost is open use that endpoint instead
	if testSockConnect("127.0.0.1", "2379") {
		config.Etcd.Endpoints = []string{"127.0.0.1:2379"}
		println("Localhost open using localhost socket instead")
	} else {
		println("Localhost NOT open using config endpoints list")
	}

	println("Client ID is: " + clientID)
	WriteToEtcd(config, keyToWrite, valueToWrite)

	readch := make(chan string)
	go readEtcdContinuously(readch, config, keyToWrite)

	for true { // loop forever (user expected to break)
		msg := <-readch
		print(msg)

		time.Sleep(3 * time.Second)
	}
}
