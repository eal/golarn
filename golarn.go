package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/leekchan/gtf"
	"github.com/thoj/go-ircevent"
	"log"
	"net/http"
	"os"
	// "text/template"
)

func handleEmpty(event map[string]interface{}, tmplString string) string {
	return fmt.Sprintf("Got empty object_kind: %s", event)
}
func handleUnknown(event map[string]interface{}, tmplString string) string {
	return fmt.Sprintf("Got unknown object_kind: %s", event["object_kind"])
}

func handleGeneric(event map[string]interface{}, tmplString string) string {
	// if tmplString == "" {
	// 	tmplString = "{{.object_kind}}: {{.}}"
	// }
	tmpl, err := gtf.New("test").Parse(tmplString)
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

func withDefaultBool(value string, fallback bool) bool {
	if value == "TRUE" {
		return true
	} else if value == "FALSE" {
		return false
	} else {
		return fallback
	}
}

func main() {
	fmt.Println("Golare har inga polare")
	nick := flag.String("nick", withDefault(os.Getenv("GOLARN_NICK"), "golarn"), "nickname")
	username := flag.String("username", withDefault(os.Getenv("GOLARN_USER"), "golarn"), "username")
	server := flag.String("server", withDefault(os.Getenv("GOLARN_SERVER"), "efnet.port80.se:6697"), "server:port")
	channel := flag.String("channel", withDefault(os.Getenv("GOLARN_CHANNEL"), "#golarn_test"), "channel to join")
	adminNick := flag.String("admin", withDefault(os.Getenv("GOLARN_ADMIN"), "someadminuser"), "admin nickname")
	password := flag.String("password", withDefault(os.Getenv("GOLARN_PASSWORD"), "t0ps3cr3t"), "password")
	part := flag.String("part", withDefault(os.Getenv("GOLARN_PART"), ""), "leave auto-joined channel on startup")
	dummy := flag.Bool("dummy", withDefaultBool(os.Getenv("GOLARN_DUMMY"), false), "dummy (don't connect to IRC, just print to stdout)")
	verbose := flag.Bool("verbose", withDefaultBool(os.Getenv("GOLARN_VERBOSE"), false), "verbose (more messages)")
	debug := flag.Bool("debug", withDefaultBool(os.Getenv("GOLARN_DEBUG"), false), "debug (more more messages)")
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
	if *dummy {
		fmt.Println("dummy: ", *dummy)
	}
	irccon := irc.IRC(*nick, *username)
	irccon.Password = *password
	irccon.VerboseCallbackHandler = *verbose
	irccon.Debug = *debug
	irccon.UseTLS = useTLS
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	irccon.AddCallback("001", func(e *irc.Event) {
		irccon.Join(*channel)
		if *part != "" {
			irccon.Part(*part)
		}
	})
	irccon.AddCallback("366", func(e *irc.Event) {})
	if !*dummy {
		err := irccon.Connect(*server)
		if err != nil {
			fmt.Printf("Err %s", err)
			return
		}

		go irccon.Loop()
	}

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
		tmplMap["build"] = withDefault(os.Getenv("GOLARN_BUILD_TEMPLATE"), "{{.object_kind}}: {{.}}")
		tmplMap["issue"] = withDefault(os.Getenv("GOLARN_ISSUE_TEMPLATE"), "{{.object_kind}}: {{.}}")
		tmplMap["merge_request"] = withDefault(os.Getenv("GOLARN_MERGE_REQUEST_TEMPLATE"), "{{.object_kind}}: {{.}}")
		tmplMap["note"] = withDefault(os.Getenv("GOLARN_NOTE_TEMPLATE"), "{{.object_kind}}: {{.}}")
		tmplMap["pipeline"] = withDefault(os.Getenv("GOLARN_PIPELINE_TEMPLATE"), "Pipeline: {{.}}")
		tmplMap["push"] = withDefault(os.Getenv("GOLARN_PUSH_TEMPLATE"), "Push from {{.user_username}} on {{.project.name}}: {{if eq (print .total_commits_count) \"1\"}} {{- (index .commits 0).message|truncatechars 50}} {{(index .commits 0).url}} {{else}} {{- .total_commits_count}} commits {{.project.web_url}}/compare/{{.before|slice 0 7}}...{{.after|slice 0 7}}{{end}}")
		tmplMap["tag_push"] = withDefault(os.Getenv("GOLARN_TAG_PUSH_TEMPLATE"), "{{.object_kind}}: {{.}}")
		tmplMap["wiki_page"] = withDefault(os.Getenv("GOLARN_WIKI_PAGE_TEMPLATE"), "{{.object_kind}}: {{.}}")
		// tmplMap["tag_push"] = ""

		lookup, ok := m["object_kind"]
		if ok {
			objectKind := fmt.Sprintf("%s", lookup)
			handleFunc, ok := handleMap[objectKind]
			if ok {
				tmpl, _ := tmplMap[objectKind]
				if !*dummy {
					irccon.Privmsg(*channel, handleFunc(m, tmpl))
				} else {
					fmt.Println(handleFunc(m, tmpl))
				}
			} else {
				if !*dummy {
					irccon.Privmsg(*adminNick, handleUnknown(m, ""))
				} else {
					fmt.Println(handleUnknown(m, ""))
				}
			}
		} else {
			if !*dummy {
				irccon.Privmsg(*adminNick, handleEmpty(m, ""))
			} else {
				fmt.Println(handleEmpty(m, ""))
			}
		}
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Golare har inga polare")
	})
	fmt.Println("Starting HTTP loop")
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("Shutting down")
}
