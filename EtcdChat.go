package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/nrm21/support"
)

var version string // to be auto-added with -ldflags at build time

// Program entry point
func main() {
	var mw *walk.MainWindow
	var configSleepMsgBox *walk.TextEdit
	var resultMsgBox *walk.TextEdit
	var chatTextBox *walk.LineEdit
	strIP := support.GetOutboundIP().String()
	clientID := generateID()
	config, err := getConfigContentsFromRegistry(`SOFTWARE\NateMorrison\EtcdChat`)
	//config, err := getConfigContentsFromYaml("support/config.yml")
	keyToWrite := fmt.Sprintf("%s/%s", config.Etcd.BaseKeyToWrite, clientID)

	// if localhost is open use that endpoint instead
	if testSockConnect("127.0.0.1", "2379") {
		config.Etcd.Endpoints = []string{"127.0.0.1:2379"}
		println("Localhost open using localhost socket instead")
	} else {
		println("Localhost NOT open using config endpoints list")
	}

	MainWindow{
		AssignTo: &mw,
		Title:    "Etcd Chat",
		Size:     Size{1024, 768},
		Layout:   VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					ScrollView{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							TextLabel{
								Text: "Sleep seconds: ",
							},
							TextEdit{
								AssignTo: &configSleepMsgBox,
								// convert int to hex before string conversion
								Text: string(config.Etcd.SleepSeconds + 0x30),
							},
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Configure",
								OnClicked: func() {
									intSleepSecs, _ := strconv.Atoi(configSleepMsgBox.Text())
									setDWordValueToRegistry(`SOFTWARE\NateMorrison\EtcdChat`, "SleepSeconds", intSleepSecs)
									config.Etcd.SleepSeconds = time.Duration(intSleepSecs)
								},
							},
						},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					TextEdit{
						AssignTo: &resultMsgBox,
						ReadOnly: true,
						MinSize:  Size{600, 610},
						Font:     Font{Family: "Ariel", PointSize: 12},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					ScrollView{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							LineEdit{AssignTo: &chatTextBox, Text: ""},
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Send",
								OnClicked: func() {
									go func() {
										timestamp := getMilliTime()
										valueToWrite := fmt.Sprintf("%s || %s . . . %s", timestamp, strIP, chatTextBox.Text()+"\r\n")
										WriteToEtcd(config, keyToWrite, valueToWrite)
										chatTextBox.SetText("")
									}()
								},
							},
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
	}.Create()

	// If the config file doesn't exist this will run from config file
	if os.IsNotExist(err) {
		walk.MsgBox(mw, "Error", "The proper config file or registry keys do not exist", walk.MsgBoxIconError)
		os.Exit(1)
	}

	// Start listening for a response
	go listenForResponse(&config, resultMsgBox, keyToWrite)

	mw.Run()
}
