package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	general "NatsChat/src/general"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	nats "github.com/nats-io/nats.go"
)

var version string // to be auto-added with -ldflags at build time
var clientID string
var config general.Config
var nc *nats.Conn

var mw *walk.MainWindow
var configIDMsgBox *walk.TextEdit
var logWindowBox *walk.TextEdit
var chatTextBox *walk.LineEdit
var showOwnMsgsCheckbox *walk.CheckBox

func sendMessage() {
	valueToWrite := ""
	// Publish message and clear textbox
	valueToWrite = fmt.Sprintf("<%s> [%s]  %s", general.GetMilliTime(),
		clientID, chatTextBox.Text()+"\r\n")
	nc.Publish(config.Nats.Subname, []byte(valueToWrite))
	chatTextBox.SetText("")
}

// Program entry point
func main() {
	var err error

	clientID = general.GenerateID()

	MainWindow{
		AssignTo: &mw,
		Title:    "Nats Chat",
		Size:     Size{1024, 768},
		Layout:   VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					ScrollView{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							TextLabel{
								Text: "Client ID: ",
							},
							TextEdit{
								AssignTo: &configIDMsgBox,
								// convert int to hex before string conversion
								Text: clientID,
							},
							PushButton{
								MinSize: Size{140, 20},
								MaxSize: Size{200, 20},
								Text:    " Generate new client ID ",
								OnClicked: func() {
									clientID = general.GenerateID()
									configIDMsgBox.SetText(clientID)
								},
							},
						},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					TextEdit{
						AssignTo: &logWindowBox,
						ReadOnly: true,
						MinSize:  Size{600, 605},
						Font:     Font{Family: "Ariel", PointSize: 12},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					ScrollView{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							LineEdit{
								AssignTo: &chatTextBox,
								OnKeyDown: func(key walk.Key) {
									if key == walk.KeyReturn {
										go sendMessage()
									}
								},
								Text: "",
							},
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Send",
								OnClicked: func() {
									go sendMessage()
								},
							},
						},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					CheckBox{
						AssignTo: &showOwnMsgsCheckbox,
						Text:     "Show own messages",
						Checked:  true,
					},
					TextLabel{ /* Text: "Client : " + clientI */ },
					TextLabel{Text: "Version : " + version},
				},
			},
		},
	}.Create()

	// Get the config from expected filepath (or args if specified)
	configPath := "../config.yml"
	if len(os.Args) > 1 {
		if os.Args[1] == "-c" { // or if program arguements
			configPath = os.Args[2]
		}
	}
	config, err = general.GetConfigContentsFromYaml(configPath)
	if err != nil { // if config file not found put an error message up and exit
		walk.MsgBox(mw, "Error", err.Error(), walk.MsgBoxIconError)
		log.Fatalln(err)
	}

	// Connect to server and subscribe
	nc, err = nats.Connect(config.Nats.Endpoints[0])
	if err != nil { // if server unavailable put an error message up and exit
		walk.MsgBox(mw, "Error", err.Error(), walk.MsgBoxIconError)
		log.Fatalln(err)
	} else {
		nc.Subscribe(config.Nats.Subname, func(msg *nats.Msg) {
			newMsg := string(msg.Data)
			sendingClient := newMsg[strings.Index(newMsg, "[")+1 : strings.Index(newMsg, "]")]
			if clientID == sendingClient { // if we are sender
				if showOwnMsgsCheckbox.Checked() == true { // if we allow seeing our own messages
					// then higlight our msgs and show
					newMsg = strings.Replace(newMsg, "[", "**[", 1)
					logWindowBox.SetText(logWindowBox.Text() + newMsg)
				}
			} else {
				// just show remote msgs
				logWindowBox.SetText(logWindowBox.Text() + newMsg)
			}
		})
	}

	mw.Run()
}
