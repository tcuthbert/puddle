package main

import (
	"./controllers"
	"github.com/perthgophers/puddle/puddle"
	"net/http"
	"os"
	"os/exec"
)

// SLACKTOKEN is the slack API token
var SLACKTOKEN string

// GITTAG Current Git Tag
var GITTAG string

func init() {
	SLACKTOKEN = os.Getenv("SLACKTOKEN")

	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		GITTAG = "NO TAG"
	} else {
		GITTAG = string(out)
	}
	GITTAG += "/"
	out, err = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil {
		GITTAG += string(out)
	}
}

func main() {
	http.HandleFunc("/logs/", controllers.LogHandler)
	http.ListenAndServe(":8080", nil)
	puddle.Run(SLACKTOKEN, GITTAG)

}
