package puddle

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/nlopes/slack"
)

var lock = new(sync.Mutex)
var adminCommands = make(map[string]func(string, string, slack.Msg) error)

// ProcessMessage performs basic routing on Slack messages in the channel #puddle
// Eventually, this will feed into two routers
// One for admin commands & one for mud commands
func ProcessMessage(msg slack.Msg) error {
	if strings.HasPrefix(msg.Text, "!") {
		username, _ := GetUsername(msg)
		ProcessAdminCommand(username, msg.Text, msg)
	}
	return nil
}

// GetUsername retrieves human readable username information from SLACK API using Slack username ID
func GetUsername(msg slack.Msg) (string, error) {
	if ISDEV == "true" {
		return "cliuser", nil
	}
	userInfo, err := slackAPI.GetUserInfo(msg.User)
	if err != nil {
		ErrorMessage("No user information: "+msg.Text, err.Error())
		return "", err
	}

	return userInfo.Name, nil
}

// ProcessAdminCommand routes admin commands, by accessing functions in a map
func ProcessAdminCommand(username string, msgText string, msg slack.Msg) {
	words := strings.Split(msgText, " ")
	cmdString := strings.TrimPrefix(words[0], "!")

	if fn, ok := adminCommands[cmdString]; ok {
		go fn(username, msgText, msg)
	} else {
		SendMessage(fmt.Sprintf("@%s Wat?", username))
	}
}

// SendMessage send slack or cli message
// shorthand for rtm.SendMessage(rtm.NewOutgoingMessage(text, CHANNEL))
func SendMessage(text string) error {
	if ISDEV == "true" {
		fmt.Println(">> ", text)
		return nil
	}
	rtm.SendMessage(rtm.NewOutgoingMessage(text, CHANNEL))

	return nil
}

// Build will pull the latest git master and rebuild. It will then restart puddlebot
// Syntax: `!build <branch-name>`
// Example: `!build master`
func Build(username string, msgText string, msg slack.Msg) error {
	SendMessage("Rebuild requested. Locked.")
	lock.Lock()

	words := strings.Split(msgText, " ")
	branch := "master"
	if len(words) > 1 {
		branch = words[1]
	}
	SendMessage(fmt.Sprintf("pulling origin/%s...", branch))

	out, err := exec.Command("git", "pull", "origin", branch).Output()
	if err != nil {
		ErrorMessage("ERROR:" + err.Error())
		return err
	}
	SendMessage(string(out))

	out, err = exec.Command("git", "checkout", branch).Output()
	if err != nil {
		SendMessage("ERROR:" + err.Error())
		return err
	}
	SendMessage(string(out))

	_, err = exec.Command("go", "install").Output()
	if err != nil {
		ErrorMessage("ERROR:" + err.Error())
		return err
	}
	SendMessage("...Restarting...")

	cmd := exec.Command("./run.sh")
	cmd.Start()

	os.Exit(1)
	return err
}

// RegisterAdminCommand Register an admin command with puddle bot
// Accepts functions in the form of `func(string, string, slack.Msg)`
func RegisterAdminCommand(cmdString string, fn func(string, string, slack.Msg) error) error {
	log.Println(fmt.Sprintf("Registering new admin command <%s>", cmdString))

	lock.Lock()
	defer lock.Unlock()

	if _, ok := adminCommands[cmdString]; !ok {
		adminCommands[cmdString] = fn
		return nil
	}

	return errors.New("Command exists")
}

func init() {
	RegisterAdminCommand("build", Build)
	RegisterAdminCommand("bitcoin", BitcoinTicker)
}