package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"net/http"
	"os"
	"text/template"
)

func handleEmpty(event map[string]interface{}, tmplString string) string {
	return fmt.Sprintf("Got empty object_kind: %s", event)
}
func handleUnknown(event map[string]interface{}, tmplString string) string {
	return fmt.Sprintf("Got unknown object_kind: %s", event["object_kind"])
}

func handlePush(event map[string]interface{}, tmplString string) string {
	if tmplString == "" {
		tmplString = "Push {{.}}"
	}
	tmpl, err := template.New("test").Parse(tmplString)
	if err != nil {
		panic(err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, event)
	if err != nil {
		panic(err)
	}
	return b.String()

	// project := event["project"].(map[string]interface{})
	// var ownCommits, othersCommits int
	// var commitMsg string
	// switch vv := event["commits"].(type) {
	// case []interface{}:
	// 	for _, commit := range vv {
	// 		c := commit.(map[string]interface{})
	// 		commitMsg = c["message"].(string)
	// 		if event["user_email"] == c["author"].(map[string]interface{})["email"] {
	// 			ownCommits++
	// 		} else {
	// 			othersCommits++
	// 		}
	// 	}
	// default:
	// 	return fmt.Sprintf("unable to parse event commits %s", event)
	// }
	// ref := event["ref"].(string)
	// var branch_msg string
	// if ref[11:] != project["default_branch"] {
	// 	branch_msg = fmt.Sprintf(" (%s)", ref[11:])
	// }

	// var totalCommits float64
	// switch v := event["total_commits_count"].(type) {
	// case float64:
	// 	totalCommits = v
	// }

	// if ownCommits > 1 || othersCommits > 0 {
	// 	return fmt.Sprintf("Push from %s on %s%s: [%d commits by self, %d commits by others, total %.0f] ",
	// 		event["user_username"],
	// 		project["name"],
	// 		branch_msg,
	// 		ownCommits,
	// 		othersCommits,
	// 		totalCommits,
	// 	)

	// } else {
	// 	return fmt.Sprintf("Push from %s on %s%s: %s",
	// 		event["user_username"],
	// 		project["name"],
	// 		branch_msg,
	// 		commitMsg,
	// 	)
	// }
}
func handleTag(event map[string]interface{}, tmplString string) string {
	return "Tage Taggelito"
}
func handleBuild(event map[string]interface{}, tmplString string) string {
	return "Byggare Bob"
}
func handleNote(event map[string]interface{}, tmplString string) string {
	return "Note this!"
}
func handleIssue(event map[string]interface{}, tmplString string) string {
	return "Issue"
}
func handleMergeRequest(event map[string]interface{}, tmplString string) string {
	return "merge request"
}
func handlePipeline(event map[string]interface{}, tmplString string) string {
	return "pipeline"
}
func handleTagPush(event map[string]interface{}, tmplString string) string {
	if tmplString == "" {
		tmplString = "tag push: {{.}}"
	}
	tmpl, err := template.New("test").Parse(tmplString)
	if err != nil {
		panic(err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, event)
	if err != nil {
		panic(err)
	}
	return b.String()
}
func handleWikiPage(event map[string]interface{}, tmplString string) string {
	return "wiki page"
}

func handleGeneric(event map[string]interface{}, tmplString string) string {
	tmpl, err := template.New("test").Parse(tmplString)
	if err != nil {
		panic(err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, event)
	if err != nil {
		panic(err)
	}
	return b.String()
}

func withDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	} else {
		return value
	}
}

func main() {
	nick := flag.String("nick", withDefault(os.Getenv("GOLARN_NICK"), "golarn"), "nickname")
	username := flag.String("username", withDefault(os.Getenv("GOLARN_USER"), "golarn"), "username")
	server := flag.String("server", withDefault(os.Getenv("GOLARN_SERVER"), "efnet.port80.se:6697"), "server:port")
	channel := flag.String("channel", withDefault(os.Getenv("GOLARN_CHANNEL"), "#golarn_test"), "channel to join")
	adminNick := flag.String("admin", withDefault(os.Getenv("GOLARN_ADMIN"), "someadminuser"), "admin nickname")
	password := flag.String("password", withDefault(os.Getenv("GOLARN_PASSWORD"), "t0ps3cr3t"), "password")
	part := flag.String("part", withDefault(os.Getenv("GOLARN_PART"), ""), "leave auto-joined channel on startup")

	tlsString := os.Getenv("GOLARN_TLS")

	// Roundabout way to set TLS option
	useTLS := true
	if tlsString != "" {
		if tlsString == "FALSE" {
			useTLS = false
		} else {
			useTLS = true
		}
	} else {
		useTLS = *flag.Bool("tls", true, "Use TLS")
	}

	flag.Parse()

	irccon := irc.IRC(*nick, *username)
	irccon.Password = *password
	irccon.VerboseCallbackHandler = true
	irccon.Debug = true
	irccon.UseTLS = useTLS
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	irccon.AddCallback("001", func(e *irc.Event) {
		irccon.Join(*channel)
		if *part != "" {
			irccon.Part(*part)
		}
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
		if t == nil {
			// got some garbage posted
			fmt.Fprintf(w, "Error: got garbage: %s\n", r.Body)
			return
		}
		m := t.(map[string]interface{})
		if err != nil {
			irccon.Privmsgf(*channel, "Got JSON decoding error %s", err)
		}
		defer r.Body.Close()
		// log.Println(t.Test)
		fmt.Fprintf(w, "OK\n")
		handleMap := map[string]func(map[string]interface{}, string) string{}
		handleMap["build"] = handleGeneric
		handleMap["issue"] = handleGeneric
		handleMap["merge_request"] = handleGeneric
		handleMap["note"] = handleGeneric
		handleMap["pipeline"] = handleGeneric
		handleMap["push"] = handleGeneric
		// handleMap["tag_push"] = handleTag
		// handleMap["tag_push"] = handleTagPush
		handleMap["tag_push"] = handleGeneric
		handleMap["wiki_page"] = handleGeneric

		tmplMap := make(map[string]string)
		tmplMap["build"] = withDefault(os.Getenv("GOLARN_BUILD_TEMPLATE"), "")
		tmplMap["issue"] = withDefault(os.Getenv("GOLARN_ISSUE_TEMPLATE"), "")
		tmplMap["merge_request"] = withDefault(os.Getenv("GOLARN_MERGE_REQUEST_TEMPLATE"), "")
		tmplMap["note"] = withDefault(os.Getenv("GOLARN_NOTE_TEMPLATE"), "")
		tmplMap["pipeline"] = withDefault(os.Getenv("GOLARN_PIPELINE_TEMPLATE"), "")
		tmplMap["push"] = withDefault(os.Getenv("GOLARN_PUSH_TEMPLATE"), "Push from {{.user_username}} on {{.project.name}}: {{if .total_commits_count==1}}{{.commits[0].comment}}")
		tmplMap["tag_push"] = withDefault(os.Getenv("GOLARN_TAG_PUSH_TEMPLATE"), "tag push: {{.}}")
		tmplMap["wiki_page"] = withDefault(os.Getenv("GOLARN_WIKI_PAGE_TEMPLATE"), "")
		// tmplMap["tag_push"] = ""

		lookup, ok := m["object_kind"]
		if ok {
			objectKind := fmt.Sprintf("%s", lookup)
			handleFunc, ok := handleMap[objectKind]
			if ok {
				tmpl, _ := tmplMap[objectKind]
				irccon.Privmsg(*channel, handleFunc(m, tmpl))
			} else {
				irccon.Privmsg(*adminNick, handleUnknown(m, ""))
			}
		} else {
			irccon.Privmsg(*adminNick, handleEmpty(m, ""))
		}
	}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
