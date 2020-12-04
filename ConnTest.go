package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"locallibs/support"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type config struct {
	// var name has to be uppercase here or it won't work
	EtcdConf etcdConfig `yaml:"etcd"`
}

type etcdConfig struct {
	// var name has to be uppercase here or it won't work
	Endpoints  []string `yaml:"endpoints"`
	KeyToWrite string   `yaml:"keyToWrite"`
	Timeout    int      `yaml:"timeout"`
	CertCa     string   `yaml:"cert-ca"`
	PeerCert   string   `yaml:"peer-cert"`
	PeerKey    string   `yaml:"peer-key"`
}

// GenerateID creates a random string 12 digits long and returns it to server as an id
func GenerateID() string {
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

// GetConfigContents Unmarshals the config contents from file into memory
func GetConfigContents(filename string) config {
	var conf config
	file := support.ReadConfigFileContents(filename)
	err := yaml.Unmarshal(file, &conf)
	if err != nil {
		fmt.Printf("There was an error decoding the yaml file. err = %s\n", err)
	}

	return conf
}

// TakeUserInput gets input from the user
func TakeUserInput() string {
	fmt.Println("Enter a message: ")
	in := bufio.NewReader(os.Stdin)
	msg, err := in.ReadString('\n')
	if err != nil {
		fmt.Printf("There was an error taking in input: %s\n", err)
	}

	return msg
}

// Main entry point
func main() {
	support.SetupCloseHandler() // setup ctrl + c to break loop
	fmt.Println("Press ctrl + c to exit...")

	strIP := support.GetOutboundIP().String()
	clientID := GenerateID()
	fmt.Println("Client ID is: " + clientID)
	config := GetConfigContents("support/config.yml")
	//message := TakeUserInput()
	message := "another test"

	for true { // loop forever (user expected to break)
		now := time.Now()
		timestamp := now.Format(time.RFC3339Nano)
		keyToWrite := config.EtcdConf.KeyToWrite + "/" + clientID
		valueToWrite := timestamp + " | " + strIP + " | " + message

		WriteToEtcd(config.EtcdConf, keyToWrite, valueToWrite)
		values := ReadFromEtcd(config.EtcdConf, config.EtcdConf.KeyToWrite)
		for k, v := range values {
			print(k + " :\n")
			print(v + "\n")
		}
		time.Sleep(3 * time.Second)
	}
}
