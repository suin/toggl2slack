//go:generate go-bindata -o bindata.go status.png
package main

import (
	"encoding/json"
	"fmt"
	"github.com/alexflint/gallium"
	"github.com/jason0x43/go-toggl"
	"github.com/mitchellh/go-homedir"
	"github.com/nlopes/slack"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"time"
)

const TOGGL_CACHE_FILE = "~/.toggl2slack/data.json"
const CONFIG_FILE = "~/.toggl2slack/config.json"
const START_MESSAGE_FORMAT = ":start: %s"
const STOP_MESSAGE_FORMAT = ":stop: %s"

func main() {
	runtime.LockOSThread()
	gallium.RedirectStdoutStderr(os.ExpandEnv("$HOME/Library/Logs/TogglToSlack.log"))
	gallium.Loop(os.Args, onReady)
}

func handleMenuQuit() {
	log.Println("quit clicked")
	os.Exit(0)
}

func onReady(app *gallium.App) {
	img, err := gallium.ImageFromPNG(MustAsset("status.png"))
	if err != nil {
		fmt.Println("unable to decode status.png:", err)
		os.Exit(1)
	}

	app.AddStatusItem(gallium.StatusItemOptions{
		Image:     img,
		Width:     22,
		Highlight: true,
		Menu: []gallium.MenuEntry{
			gallium.MenuItem{Title: "Quit Toggl2Slack", OnClick: handleMenuQuit},
		},
	})

	go func() {
		for {
			config, err := loadConfig(CONFIG_FILE)
			if err != nil {
				log.Printf("%s\n", err)
				return
			}

			err = togglToSlack(config)
			if err != nil {
				log.Printf("%s\n", err)
			}
			time.Sleep(20 * time.Second)
		}
	}()
}

func togglToSlack(config Config) error {
	log.Println("Fetch toggl time entries")
	entries, err := getTimeEntries(config.TogglToken)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err = notifyStart(entry, config.SlackToken, config.SlackChannel); err != nil {
			return err
		}
		if err = notifyStop(entry, config.SlackToken, config.SlackChannel); err != nil {
			return err
		}
	}
	return nil
}

type Config struct {
	SlackToken   string
	SlackChannel string
	TogglToken   string
}

func loadConfig(filename string) (config Config, err error) {
	file, err := homedir.Expand(filename)
	if err != nil {
		return config, err
	}
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(contents, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

type State struct {
	Start []int
	Stop  []int
}

func (state *State) IsStartNotified(id int) bool {
	for _, s := range state.Start {
		if s == id {
			return true
		}
	}
	return false
}

func (state *State) IsStopNotified(id int) bool {
	for _, s := range state.Stop {
		if s == id {
			return true
		}
	}
	return false
}

func (state *State) NotifyStart(id int) {
	state.Start = append(state.Start, id)
}

func (state *State) NotifyStop(id int) {
	state.Stop = append(state.Stop, id)
}

func notifyStart(entry toggl.TimeEntry, slackToken, channel string) error {
	state, err := loadState()
	if err != nil {
		return err
	}
	if state.IsStartNotified(entry.ID) == false {
		log.Println("Post start message")
		err = postMessage(slackToken, channel, fmt.Sprintf(START_MESSAGE_FORMAT, entry.Description))
		if err != nil {
			return err
		}
		state.NotifyStart(entry.ID)
	}
	return saveState(state)
}

func notifyStop(entry toggl.TimeEntry, slackToken, channel string) error {
	state, err := loadState()
	if err != nil {
		return err
	}
	if entry.Duration < 0 {
		return nil
	}
	if state.IsStopNotified(entry.ID) == false {
		log.Println("Post stop message")
		err = postMessage(slackToken, channel, fmt.Sprintf(STOP_MESSAGE_FORMAT, entry.Description))
		if err != nil {
			return err
		}
		state.NotifyStop(entry.ID)
	}
	return saveState(state)
}

func saveState(state State) error {
	cache, err := homedir.Expand(TOGGL_CACHE_FILE)
	if err != nil {
		return err
	}
	contents, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cache, contents, os.ModePerm)
}

func loadState() (state State, err error) {
	cache, err := homedir.Expand(TOGGL_CACHE_FILE)
	if err != nil {
		return state, err
	}
	_, err = os.Stat(cache)
	if err != nil {
		return state, nil
	}
	contents, err := ioutil.ReadFile(cache)
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(contents, &state)
	if err != nil {
		return state, err
	}
	return state, nil
}

func getTimeEntries(token string) ([]toggl.TimeEntry, error) {
	session := toggl.OpenSession(token)
	account, err := session.GetAccount()
	if err != nil {
		return []toggl.TimeEntry{}, err
	}
	return account.Data.TimeEntries, nil
}

func postMessage(token, channel, message string) error {
	api := slack.New(token)
	_, _, err := api.PostMessage(channel, message, slack.PostMessageParameters{})
	return err
}
