package general

import (
	"regexp"
	"testing"
)

func TestGenerateID(t *testing.T) {
	recieved := GenerateID()
	matchesExpected, _ := regexp.MatchString("^[0-9a-z]{8}$", recieved)
	if matchesExpected == false {
		t.Errorf("Recieved: %q  Expecting: 8 digit alphanumeric", recieved)
	}
}
