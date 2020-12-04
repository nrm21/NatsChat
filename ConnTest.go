package main

import (
	"crypto/rand"
	"fmt"
	"locallibs/support"
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

// Main entry point
func main() {
	support.SetupCloseHandler() // setup ctrl + c to break loop
	println("Press ctrl + c to exit...")

	strIP := support.GetOutboundIP().String()
	clientID := GenerateID()
	// fmt.Println("Enter a message: ")
	// var msg string
	// fmt.Scanf("%q", &msg)

	var config config
	err := yaml.Unmarshal(support.ReadConfigFileContents("support/config.yml"), &config)
	if err != nil {
		fmt.Printf("There was an error decoding the yaml file. err = %s\n", err)
		return
	}

	var (
		now       time.Time
		timestamp string
	)
	for true { // loop forever (user expected to break)
		now = time.Now()
		timestamp = now.Format(time.RFC3339Nano)
		keyToWrite := config.EtcdConf.KeyToWrite + "/" + clientID
		valueToWrite := timestamp + " | " + strIP + " | " //+ " | " + msg

		WriteToEtcd(config.EtcdConf, keyToWrite, valueToWrite)
		values := ReadFromEtcd(config.EtcdConf, keyToWrite)
		for _, value := range values {
			print(value + "\n")
		}
		time.Sleep(3 * time.Second)
	}
}
