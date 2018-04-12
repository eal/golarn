package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"net/http"
)

func handleEmpty(event map[string]interface{}) string {
	return fmt.Sprintf("Got empty object_kind: %s", event)
}
func handleUnknown(event map[string]interface{}) string {
	return fmt.Sprintf("Got unknown object_kind: %s", event)
}

func handlePush(event map[string]interface{}) string {
	project := event["project"].(map[string]interface{})
	var ownCommits, othersCommits int
	var commitMsg string
	switch vv := event["commits"].(type) {
	case []interface{}:
		for _, commit := range vv {
			c := commit.(map[string]interface{})
			commitMsg = c["message"].(string)
			if event["user_email"] == c["author"].(map[string]interface{})["email"] {
				ownCommits++
			} else {
				othersCommits++
			}
		}
	default:
		return fmt.Sprintf("unable to parse event commits %s", event)
	}
	ref := event["ref"].(string)
	var branch_msg string
	if ref[11:] != project["default_branch"] {
		branch_msg = fmt.Sprintf(" (%s)", ref[11:])
	}

	var totalCommits float64
	switch v := event["total_commits_count"].(type) {
	case float64:
		totalCommits = v
	}

	if ownCommits > 1 || othersCommits > 0 {
		return fmt.Sprintf("Push from %s on %s%s: [%d commits by self, %d commits by others, total %.0f] ",
			event["user_username"],
			project["name"],
			branch_msg,
			ownCommits,
			othersCommits,
			totalCommits,
		)

	} else {
		return fmt.Sprintf("Push from %s on %s%s: %s",
			event["user_username"],
			project["name"],
			branch_msg,
			commitMsg,
		)
	}
}
func handleTag(event map[string]interface{}) string {
	return "Tage Taggelito"
}

func main() {
	nick := flag.String("nick", "goltupp", "nickname")
	username := flag.String("username", *nick, "username")
	useTLS := flag.Bool("tls", true, "Use TLS")
	server := flag.String("server", "efnet.port80.se:6697", "server:port")
	channel := flag.String("channel", "#goltupp_test", "channel to join")
	adminNick := flag.String("admin", "someadminuser", "admin nickname")
	password := flag.String("password", "t0ps3cr3t", "password")
	part := flag.String("part", "#notmychannel", "leave auto-joined channel on startup")

	flag.Parse()

	irccon := irc.IRC(*nick, *username)
	irccon.Password = *password
	irccon.VerboseCallbackHandler = true
	irccon.Debug = true
	irccon.UseTLS = *useTLS
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	irccon.AddCallback("001", func(e *irc.Event) {
		irccon.Join(*channel)
		irccon.Part(*part)
	})
	irccon.AddCallback("366", func(e *irc.Event) {})
	err := irccon.Connect(*server)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	go irccon.Loop()

	handler := func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var t interface{}
		err := decoder.Decode(&t)
		m := t.(map[string]interface{})
		if err != nil {
			irccon.Privmsgf(*channel, "Got JSON decoding error %s", err)
		}
		defer r.Body.Close()
		// log.Println(t.Test)
		fmt.Fprintf(w, "OK\n")
		handleMap := map[string]func(map[string]interface{}) string{}
		handleMap["push"] = handlePush
		handleMap["tag_push"] = handleTag
		lookup, ok := m["object_kind"]
		if ok {
			objectKind := fmt.Sprintf("%s", lookup)
			handleFunc, ok := handleMap[objectKind]
			if ok {

				irccon.Privmsg(*channel, handleFunc(m))
			} else {
				irccon.Privmsg(*adminNick, handleUnknown(m))
			}
		} else {
			irccon.Privmsg(*adminNick, handleEmpty(m))
		}
	}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
