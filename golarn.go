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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"strings"
	"text/template"
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
	funcMap := gtf.GtfTextFuncMap
	funcMap["joinlines"] = func(s string) string {
		//fl := strings.Split(s, "\n")[0]
		fl := strings.Replace(s, "\n", "|", -1)
		return fl
	}
	tmpl, err := template.New("test").Funcs(funcMap).Parse(tmplString)
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

func jsonMap(input io.Reader) (map[string]interface{}, error) {
	decoder := json.NewDecoder(input)
	var t interface{}
	err := decoder.Decode(&t)
	if t == nil {
		return make(map[string]interface{}), fmt.Errorf("error decoding json")
	}
	ret := t.(map[string]interface{})
	if err != nil {
		return make(map[string]interface{}), fmt.Errorf("error transforming json")
	}
	return ret, nil
}

func withDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func withDefaultBool(value string, fallback bool) bool {
	if strings.ToLower(value) == "true" {
		return true
	} else if strings.ToLower(value) == "false" {
		return false
	} else {
		return fallback
	}
}

func verifyChannel(targetChannel string, channels string) bool {
	targetOk := false
	for _, c := range strings.Split(channels, " ") {
		if c == targetChannel {
			targetOk = true
		}
	}
	return targetOk
}

func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK, healthy")
}

func fallback(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Golare har inga polare")
}

func main() {
	fmt.Println("Golare har inga polare")
	nick := flag.String("nick", withDefault(os.Getenv("GOLARN_NICK"), withDefault(os.Getenv("HOSTNAME"), "golarn")), "nickname")
	username := flag.String("username", withDefault(os.Getenv("GOLARN_USER"), withDefault(os.Getenv("HOSTNAME"), "golarn")), "username")
	server := flag.String("server", withDefault(os.Getenv("GOLARN_SERVER"), "efnet.port80.se:6697"), "server:port")
	channels := flag.String("channels", withDefault(os.Getenv("GOLARN_CHANNELS"), "#golarn_test"), "channels to join/whitelist of nicknames")
	adminNick := flag.String("admin", withDefault(os.Getenv("GOLARN_ADMIN"), "someadminuser"), "admin nickname")
	password := flag.String("password", withDefault(os.Getenv("GOLARN_PASSWORD"), "t0ps3cr3t"), "password")
	part := flag.String("part", withDefault(os.Getenv("GOLARN_PART"), ""), "leave auto-joined channel on startup")
	dummy := flag.Bool("dummy", withDefaultBool(os.Getenv("GOLARN_DUMMY"), false), "dummy (don't connect to IRC, just print to stdout)")
	verbose := flag.Bool("verbose", withDefaultBool(os.Getenv("GOLARN_VERBOSE"), false), "verbose (more messages)")
	debug := flag.Bool("debug", withDefaultBool(os.Getenv("GOLARN_DEBUG"), false), "debug (more more messages)")
	useTLS := flag.Bool("tls", withDefaultBool(os.Getenv("GOLARN_TLS"), true), "use TLS")

	flag.Parse()
	if *dummy {
		fmt.Println("dummy: ", *dummy)
	}
	irccon := irc.IRC(*nick, *username)
	irccon.Password = *password
	irccon.VerboseCallbackHandler = *verbose
	irccon.Debug = *debug
	irccon.UseTLS = *useTLS
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	irccon.AddCallback("001", func(e *irc.Event) {
		for _, channel := range strings.Split(*channels, " ") {
			if channel[0] == '#' {
				irccon.Join(channel)
			}
			if *part != "" {
				irccon.Part(*part)
			}
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

	webhookHandler := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		// get channel
		p := strings.Split(r.URL.Path, "/")
		if len(p) != 3 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "couldn't parse URL, %v %d\n", p, len(p))
			return
		}
		var targetChannel string
		if len(p[2]) != 0 {
			// We got a channel name
			if p[2][0] == '@' {
				// directed at user
				targetChannel = p[2][1:]
			} else {
				// to channel, prepend #
				targetChannel = fmt.Sprintf("#%v", p[2])
			}
			if !verifyChannel(targetChannel, *channels) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Channel %s not found in valid channels (%s)", targetChannel, *channels)
				return
			}

		} else {
			// no target given, use the first as default
			targetChannel = strings.Split(*channels, " ")[0]
		}

		m, err := jsonMap(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			irccon.Privmsgf(targetChannel, "Got JSON decoding error %s", err)
			fmt.Fprintf(w, "Panic!\n")
			return
		}

		// log.Println(t.Test)
		fmt.Fprintf(w, "OK\n")

		lookup, ok := m["object_kind"]
		if ok {
			objectKind := fmt.Sprintf("%s", lookup)
			envString := fmt.Sprintf("GOLARN_TEMPLATE_%s", strings.ToUpper(objectKind))
			//handleFunc, ok := handleMap[objectKind]
			tmpl := os.Getenv(envString)
			if tmpl != "" {
				if !*dummy {
					irccon.Privmsg(targetChannel, handleGeneric(m, tmpl))
				} else {
					fmt.Printf(">>> %v\n", targetChannel)
					fmt.Println(handleGeneric(m, tmpl))
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

	http.HandleFunc("/webhook/", webhookHandler)
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/", fallback)
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Starting HTTP loop")
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("Shutting down")
}
