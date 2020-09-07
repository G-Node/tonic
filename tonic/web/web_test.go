package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestWebPlain(t *testing.T) {
	srv := New(4242)

	srv.Start()
	defer srv.Stop()
}

func TestWebWithRoutes(t *testing.T) {
	srv := New(4242)

	router := srv.Router
	router.StrictSlash(true)

	testget := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("(get) hello"))
	}

	testpost := func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			t.Fatalf("Post request handler failed to read form data: %v", err.Error())
		}
		resp := r.PostForm.Get("response")
		w.Write([]byte(fmt.Sprintf("(post) hello: %s", resp)))
	}

	router.HandleFunc("/test", testget).Methods("GET")
	router.HandleFunc("/test", testpost).Methods("POST")

	srv.Start()
	defer srv.Stop()

	if resp, err := http.Get("http://localhost:4242/test"); err != nil {
		t.Fatalf("Error testing get request: %v", err.Error())
	} else if b, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fatalf("Error reading get request body: %v", err.Error())
	} else if string(b) != "(get) hello" {
		t.Fatalf("Got unexpected response from get request: %s", string(b))
	}

	if resp, err := http.PostForm("http://localhost:4242/test", url.Values{"response": {"formvalue"}}); err != nil {
		t.Fatalf("Error testing post request: %v", err.Error())
	} else if b, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fatalf("Error reading post request body: %v", err.Error())
	} else if string(b) != "(post) hello: formvalue" {
		t.Fatalf("Got unexpected response from post request: %s", string(b))
	}
}

func TestErrorResponse(t *testing.T) {
	srv := New(4242)

	router := srv.Router
	router.StrictSlash(true)

	expresp := "TESTING:UNAUTHORISED"
	testget := func(w http.ResponseWriter, r *http.Request) {
		srv.ErrorResponse(w, http.StatusUnauthorized, expresp)
	}
	router.HandleFunc("/test", testget).Methods("GET")
	srv.Start()
	defer srv.Stop()

	if resp, err := http.Get("http://localhost:4242/test"); err != nil {
		t.Fatalf("Error testing get request: %v", err.Error())
	} else if b, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fatalf("Error reading get request body: %v", err.Error())
	} else if !strings.Contains(string(b), expresp) {
		t.Fatalf("Got unexpected response from get request: %s", string(b))
	}
}
