package puddle

import (
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"os"
)

// SLACKTOKEN is the slack API token
var SLACKTOKEN string

//Current Git Tag
var GITTAG string

// CHANNEL is the Slack ID for channel #puddle
var CHANNEL string = "C32K3QDFE"

var ISDEV string = os.Getenv("PUDDLEDEV")

var rtm *slack.RTM
var slackAPI *slack.Client

//Run main Puddle process, defau
func Run(token, gittag string) {

	if ISDEV == "true" || SLACKTOKEN == "" {
		ISDEV = "true"
		RunCLI()
		return
	}

	SLACKTOKEN = token
	GITTAG = gittag

	slackAPI = slack.New(SLACKTOKEN)

	// If you set debugging, it will log all requests to the console
	// Useful when encountering issues
	slackAPI.SetDebug(true)

	rtm = slackAPI.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			log.Print("Event Received: ")
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				// Ignore hello

			case *slack.ConnectedEvent:
				log.Println("######### Connected to Slack #########")
				// log.Println("Infos:", ev.Info)
				// log.Println("Connection counter:", ev.ConnectionCount)
				// Replace #general with your Channel ID
				rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("... and I'm back! Git tag: %s", GITTAG), CHANNEL))

			case *slack.MessageEvent:
				j, _ := json.Marshal(ev.Msg)
				log.Printf("Message: %v\n", string(j))
				ProcessMessage(ev.Msg)

			case *slack.PresenceChangeEvent:
				log.Printf("Presence Change: %v\n", ev)

			case *slack.LatencyReport:
				log.Printf("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				log.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				log.Printf("Invalid credentials")
				break Loop

			default:

				// Ignore other events..
				// log.Printf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}