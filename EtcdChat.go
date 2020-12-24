/*   TODO:
.	1) Message on startup if there is no support dir or files
.	2) Start the continuous etcd read on startup instead of after button press
*/

package main

import (
	"fmt"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/nrm21/support"
)

var version string // to be auto-added with -ldflags at build time

// Program entry point
func main() {
	//regKey := `SOFTWARE\NateMorrison\CalcBandwidth`
	//regValue := "bwCurrentUsed"

	// k, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE)
	// if err != nil {
	// 	// key doesn't exist lets create it
	// 	k, _, err = registry.CreateKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE|registry.SET_VALUE)
	// 	if err != nil {
	// 		log.Fatalf("Error creating registry key, exiting program")
	// 	}
	// 	// then write current value to the key
	// 	k, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE|registry.SET_VALUE)
	// 	if err != nil {
	// 		log.Fatalf("Unable to open key to write too")
	// 	} else {
	// 		k.SetStringValue(regValue, strconv.FormatFloat(bwCurrentUsed, 'f', -1, 64))
	// 	}
	// }
	// s, _, err := k.GetStringValue(regValue)

	var resultMsgBox *walk.TextEdit
	var chatTextBox *walk.LineEdit
	strIP := support.GetOutboundIP().String()
	clientID := generateID()
	config := getConfigContents("support/config.yml")

	// if localhost is open use that endpoint instead
	if testSockConnect("127.0.0.1", "2379") {
		config.Etcd.Endpoints = []string{"127.0.0.1:2379"}
		println("Localhost open using localhost socket instead")
	} else {
		println("Localhost NOT open using config endpoints list")
	}

	MainWindow{
		Title:  "Etcd Chat",
		Size:   Size{1024, 768},
		Layout: VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					TextEdit{
						AssignTo: &resultMsgBox,
						ReadOnly: true,
						Font:     Font{Family: "Ariel", PointSize: 12},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					LineEdit{AssignTo: &chatTextBox, Text: ""},
					PushButton{
						MaxSize: Size{30, 20},
						Text:    "Send",
						OnClicked: func() {
							go func() {
								timestamp := getMicroTime()
								keyToWrite := fmt.Sprintf("%s/%s", config.Etcd.BaseKeyToWrite, clientID)
								valueToWrite := fmt.Sprintf("%s ___ %s ___ %s", timestamp, strIP, chatTextBox.Text()+"\r\n")
								WriteToEtcd(config, keyToWrite, valueToWrite)

								readch := make(chan string)
								go readEtcdContinuously(readch, config, keyToWrite)

								for true { // loop forever (user expected to break)
									msg := <-readch
									msg = resultMsgBox.Text() + msg
									resultMsgBox.SetText(msg)
									time.Sleep(config.Etcd.SleepSeconds * time.Second)
								}
							}()
						},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					TextLabel{Text: strIP + "   |   " + clientID},
					TextLabel{Text: "Version: " + version},
				},
			},
		},
	}.Run()
}
