package main

import (
	"encoding/json"
	// "fmt"
	"os"
	"strings"
	"testing"
)

func getEvent(eventFile string) map[string]interface{} {
	f, err := os.Open(eventFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	var t interface{}
	err = decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	if t == nil {
		panic("couldn't decode JSON")
	}
	return t.(map[string]interface{})
}

func TestDummy(t *testing.T) {
	return
}

func TestHandleEmpty(t *testing.T) {
	empty := getEvent("events/broken/no_object_kind.json")
	e := handleEmpty(empty, "someignoredtemplatestring")
	// fmt.Println(e)
	if !strings.Contains(e, "empty object_kind") {
		t.Fatalf("could not handle empty test case: %s", e)
	}
}

func TestHandleUnknown(t *testing.T) {
	unknown := getEvent("events/broken/no_object_kind.json")
	e := handleUnknown(unknown, "someignoredtemplatestring")
	// fmt.Println(e)
	if !strings.Contains(e, "unknown object_kind") {
		t.Fatalf("could not handle unknown test case: %s", e)
	}
}
