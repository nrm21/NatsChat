package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/lxn/walk"
	"github.com/nrm21/support"
	"golang.org/x/sys/windows/registry"
	"gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	Etcd struct {
		// var name has to be uppercase here or it won't work
		Endpoints      []string      `yaml:"endpoints"`
		BaseKeyToWrite string        `yaml:"baseKeyToWrite"`
		Timeout        int           `yaml:"timeout"`
		SleepSeconds   time.Duration `yaml:"sleepSeconds"`
		CertCa         string        `yaml:"cert-ca"`
		PeerCert       string        `yaml:"peer-cert"`
		PeerKey        string        `yaml:"peer-key"`
	}
}

// Unmarshals the config contents from file into memory
func getConfigContentsFromYaml(filename string) (Config, error) {
	var conf Config
	file, err := support.ReadConfigFileContents(filename)
	if err != nil {
		fmt.Printf("The file was not found. err = %s\n", err)
		return conf, err
	}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		fmt.Printf("There was an error decoding the yaml file. err = %s\n", err)
		return conf, err
	}

	return conf, nil
}

// Gets config values from registry keys
func getConfigContentsFromRegistry(regKey string) (Config, error) {
	var conf Config

	k, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE)
	if err != nil {
		return conf, err // the reg key doesn't exist
	}

	// Now return contents of specified keys
	conf.Etcd.Endpoints, _, err = k.GetStringsValue("Endpoints")
	conf.Etcd.BaseKeyToWrite, _, err = k.GetStringValue("BaseKeyToWrite")
	conf.Etcd.CertCa, _, err = k.GetStringValue("CertCa")
	conf.Etcd.PeerCert, _, err = k.GetStringValue("PeerCert")
	conf.Etcd.PeerKey, _, err = k.GetStringValue("PeerKey")
	if err != nil {
		return conf, err // a reg value may be missing
	}
	timeout, _, err := k.GetIntegerValue("Timeout")
	conf.Etcd.Timeout = int(timeout)
	if err != nil {
		return conf, err // a reg value may be missing
	}
	sleepsecs, _, err := k.GetIntegerValue("SleepSeconds")
	conf.Etcd.SleepSeconds = time.Duration(sleepsecs)
	if err != nil {
		return conf, err // a reg value may be missing
	}

	return conf, nil
}

// Saves the given value path to the registry
func setDWordValueToRegistry(regKey string, regValue string, sleepSecs int) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return err // the reg key doesn't exist
	}
	err = k.SetDWordValue(regValue, uint32(sleepSecs))
	if err != nil {
		return err // something else went wrong
	}

	return nil
}

// Creates a random string 12 digits long and returns it to server as an id
func generateID() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	for i := range bytes {
		n := int(bytes[i]) // convert to int
		/* now mod our number by 36 so we get a result bewteen 1 and 35, anything less than
		   10 is assigned a digit, and anything greater is assigned a lowercase letter.
		   48 decimal is 0x30 or ascii '0' and 97 decimal is 0x61 or ascii 'a'. */
		n = (n % 36)
		if n < 10 {
			n += 0x30
		} else {
			n += 0x57 // ten less than 0x61 to account for shift
		}
		bytes[i] = byte(n) // now convert back to byte
	}

	return string(bytes)
}

// Gets input from the user to store in a var
func takeUserInput() string {
	fmt.Println("Enter a message: ")
	in := bufio.NewReader(os.Stdin)
	msg, err := in.ReadString('\n')
	if err != nil {
		fmt.Printf("There was an error taking in input: %s\n", err)
	}

	return msg
}

// Returns a string of (up to) the nanosecond level of right now (at runtime)
func getMilliTime() string {
	now := time.Now()
	tstamp := now.Format(time.RFC3339Nano)
	tstamp = strings.Replace(tstamp, "T", "  ", 1)

	return tstamp[:len(tstamp)-14] // second resolution
}

// Continuously prints read variables to screen except the ones we wrote
func readEtcdContinuously(readch chan string, config *Config, keyWritten string) {
	for true {
		values := ReadFromEtcd(*config, config.Etcd.BaseKeyToWrite)

		// remove the value we wrote ourselves
		if _, ok := values[keyWritten]; ok {
			delete(values, keyWritten)
		}

		// now put them in a string for printing
		msg := ""
		for _, v := range values {
			//msg += k + " :\t"	// key reading from
			msg += v + "\n" // timestamp + IP + msg
		}

		// then delete printed key from etcd
		for keyToDelete := range values {
			DeleteFromEtcd(*config, keyToDelete)
		}

		// and send into channel
		readch <- msg
	}
}

// Checks a socket connection and returns bool of if open or not
func testSockConnect(host string, port string) bool {
	conn, _ := net.DialTimeout("tcp", net.JoinHostPort(host, port), 500*time.Millisecond)
	if conn != nil {
		defer conn.Close()

		return true
	} else {
		return false
	}
}

// loops forever waiting for a response to the channel
func listenForResponse(config *Config, resultMsgBox *walk.TextEdit, keyToWrite string) {
	readch := make(chan string)

	// This needs its own thread since it also loops forever
	go readEtcdContinuously(readch, config, keyToWrite)

	for true { // loop forever (user expected to break)
		msg := <-readch
		msg = resultMsgBox.Text() + msg
		resultMsgBox.SetText(msg)
		time.Sleep(config.Etcd.SleepSeconds * time.Second)
	}
}
