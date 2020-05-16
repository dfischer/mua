package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/neilalexander/mdl/src/mdl"
)

var (
	Black   = "\033[1;30m%s\033[0m"
	Red     = "\033[1;31m%s\033[0m"
	Green   = "\033[1;32m%s\033[0m"
	Yellow  = "\033[1;33m%s\033[0m"
	Purple  = "\033[1;34m%s\033[0m"
	Magenta = "\033[1;35m%s\033[0m"
	Teal    = "\033[1;36m%s\033[0m"
	White   = "\033[1;37m%s\033[0m"
)

var hsURL = *flag.String("hsurl", "https://matrix.org", "the homeserver URL to connect to")
var userID = *flag.String("user", "", "the user ID to connect with, or blank for guest")
var accessToken = *flag.String("accesstoken", "", "the access token to connect with, or blank for guest")
var roomID = *flag.String("room", "!TdSVXZoEcLugVpglQn:matrix.org", "the room ID to use as our environment")

var encode = flag.String("encode", "", "encode the given file and print out event content")

func main() {
	flag.Parse()

	if encode != nil && *encode != "" {
		file, err := ioutil.ReadFile(*encode)
		if err != nil {
			panic(err)
		}
		src := mdl.Source{
			Type:   mdl.SourceTypeLua,
			Source: mdl.SourceCode(file),
		}
		j, err := json.MarshalIndent(src, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(j))
		return
	}

	if homedir, err := os.UserHomeDir(); err == nil {
		if file, err := ioutil.ReadFile(homedir + "/.mdlconfig"); err == nil {
			var cookie map[string]string
			if err = json.Unmarshal(file, &cookie); err == nil {
				hsURL = cookie["home_server"]
				userID = cookie["user_id"]
				accessToken = cookie["access_token"]
			}
		}
	}

	client, err := mdl.NewClient(hsURL, userID, accessToken)
	if err != nil {
		panic(err)
	}

	if homedir, err := os.UserHomeDir(); err == nil {
		cookie := struct {
			HSURL       string `json:"home_server"`
			UserID      string `json:"user_id"`
			AccessToken string `json:"access_token"`
		}{
			HSURL:       hsURL,
			UserID:      client.UserID(),
			AccessToken: client.AccessToken(),
		}
		j, err := json.MarshalIndent(cookie, "", "  ")
		if err == nil {
			_ = ioutil.WriteFile(homedir+"/.mdlconfig", j, 0600)
		}
	}

	room, err := client.NewRoom(roomID)
	if err != nil {
		panic(err)
	}
	lua, err := room.NewLua()
	if err != nil {
		panic(err)
	}

	if err := lua.Execute(`
		load("main")
	`); err != nil {
		fmt.Printf(Teal+"\n", "lua error: "+err.Error())
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(Yellow, ">> ")
		cmd, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		switch strings.Trim(cmd, "\t\n\r ") {
		case "exit":
			return
		default:
			if err := lua.Execute(cmd); err != nil {
				fmt.Printf(Red+"\n", "lua error: "+err.Error())
			}
		}
	}
}
