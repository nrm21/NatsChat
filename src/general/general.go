package general

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/nrm21/support"
	"gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	Nats struct {
		// var name has to be uppercase here or it won't work
		Endpoints []string `yaml:"endpoints"`
		Timeout   int      `yaml:"timeout"`
		Subname   string   `yaml:"subname"`
	}
}

// Unmarshals the config contents from file into memory
func GetConfigContentsFromYaml(filename string) (Config, error) {
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

// Creates a random string 8 digits long and returns it to server as an id
func GenerateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	for i := range bytes {
		n := int(bytes[i]) // convert to int
		/* now mod our number by 36 so we get a result bewteen 0 and 35, anything less than
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

// Returns a string of (up to) the nanosecond level of right now (at runtime)
func GetMilliTime() string {
	now := time.Now()
	tstamp := now.Format(time.RFC3339Nano)
	tstamp = strings.Replace(tstamp, "T", " ", 1)

	return tstamp[:19] // second resolution
}
