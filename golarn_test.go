package main

import (
	"encoding/json"
	// "fmt"
	"net/http"
	"net/http/httptest"
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

func TestJsonMap(t *testing.T) {
	f, err := os.Open("events/push.json")
	if err != nil {
		t.Fatalf("%s", err)
	}
	m, err := jsonMap(f)
	if err != nil {
		t.Fatalf("%s", err)
	}
	n, ok := m["object_kind"]
	if !ok {
		t.Fatalf("couldn't parse json file")
	}
	if n != "push" {
		t.Fatalf("wrong lookup ('push' != '%s')", n)
	}

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

func TestHandleGeneric1(t *testing.T) {
	generic := getEvent("events/push.json")
	e := handleGeneric(generic, "{{.object_kind}}")
	// fmt.Println(e)
	if !strings.Contains(e, "push") {
		t.Fatalf("could not handle generic test case: %s", e)
	}
}

func TestHandleGenericEscape(t *testing.T) {
	generic := getEvent("events/push.json")
	generic["object_kind"] = "foo & bar"
	e := handleGeneric(generic, "{{.object_kind}}")
	// fmt.Println(e)
	if !strings.Contains(e, "foo & bar") {
		t.Fatalf("could not handle html escape test case: %s", e)
	}
}

func TestHandleGeneric2(t *testing.T) {
	generic := getEvent("events/push.json")
	e := handleGeneric(generic, "Push from {{.user_username}} on {{.project.name}}: {{if eq (print .total_commits_count) \"1\"}} {{- (index .commits 0).message|truncatechars 50}} {{(index .commits 0).url}} {{else}} {{- .total_commits_count}} commits {{.project.web_url}}/compare/{{.before|slice 0 7}}...{{.after|slice 0 7}}{{end}}")
	// fmt.Println(e)
	if !strings.Contains(e, "Push from jsmith on Diaspora: Update Catalan translation to e38cb41. http://example.com/mike/diaspora/commit/b6568db1bc1dcd7f8b4d5a946b0b91f9dacd7327") {
		t.Fatalf("could not handle generic test case: %s", e)
	}
}

func TestWithDefault(t *testing.T) {
	if withDefault("", "foo") != "foo" {
		t.Fatalf("withDefault doesn't return default")
	} else if withDefault("foo", "bar") != "foo" {
		t.Fatalf("withDefault returns default when it shouldn't")
	}
}

func TestWithDefaultBool(t *testing.T) {
	if withDefaultBool("", true) != true {
		t.Fatalf("withDefault doesn't return true default")
	} else if withDefaultBool("", false) != false {
		t.Fatalf("withDefault doesn't return false default")
	} else if withDefaultBool("True", false) != true {
		t.Fatalf("withDefault returns default when it shouldn't (1)")
	} else if withDefaultBool("FALSE", true) != false {
		t.Fatalf("withDefault returns default when it shouldn't (2)")
	}
}
func TestVerifyChannel(t *testing.T) {
	if !verifyChannel("#foo", "#this #that #foo #bar someone") {
		t.Fatalf("couldn't verify channel")
	}
	if verifyChannel("#foo", "#this #that #bar someone") {
		t.Fatalf("erroneously verified invalid channel")
	}

}

func TestHealthz(t *testing.T) {
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthz)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `Golare har inga polare`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got '%v' want %v",
			rr.Body.String(), expected)
	}
}
