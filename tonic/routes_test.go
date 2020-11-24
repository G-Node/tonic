package tonic

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/G-Node/tonic/tonic/db"
	"github.com/G-Node/tonic/tonic/form"
)

func TestLoginRedirect(t *testing.T) {
	f := new(form.Form)
	f.Pages = []form.Page{{Elements: make([]form.Element, 1)}}
	srv, err := NewService(*f, nil, echoAction, Config{CookieName: "test-cookie"})
	if err != nil {
		t.Fatalf("failed to initialise tonic service: %s", err.Error())
	}
	handler := srv.web.Handler

	checkStatusNoCookie := func(method, route string) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest(method, route, nil)
		if err != nil {
			t.Errorf("failed to create request: %s %s", method, route)
		}
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusFound {
			t.Errorf("handler returned wrong status code: got %v expected %v", status, http.StatusFound)
		}
	}

	checkStatusNoCookie("GET", "/")
	checkStatusNoCookie("POST", "/")
	checkStatusNoCookie("GET", "/log")
	checkStatusNoCookie("GET", "/log/42")
	checkStatusNoCookie("GET", "/log/1337")

	checkStatusBadCookie := func(method, route string) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest(method, route, nil)
		if err != nil {
			t.Errorf("failed to create request: %s %s", method, route)
		}
		req.Header.Add("Cookie", "test-cookie=bad")
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusFound {
			t.Errorf("handler returned wrong status code: got %v expected %v", status, http.StatusFound)
		}
	}

	checkStatusBadCookie("GET", "/")
	checkStatusBadCookie("POST", "/")
	checkStatusBadCookie("GET", "/log")
	checkStatusBadCookie("GET", "/log/42")
	checkStatusBadCookie("GET", "/log/1337")
}

func TestFormRoutes(t *testing.T) {
	f := new(form.Form)
	f.Pages = []form.Page{{Elements: make([]form.Element, 1)}}
	srv, err := NewService(*f, nil, echoAction, Config{CookieName: "test-cookie"})
	if err != nil {
		t.Fatalf("failed to initialise tonic service: %s", err.Error())
	}
	handler := srv.web.Handler

	// Add test cookie to the database
	testSession := db.NewSession("test-token", 198)
	srv.db.InsertSession(testSession)

	cookie := testSession.ID

	// Load the form
	getReq, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error("failed to create request: GET /")
	}
	getReq.Header.Add("Cookie", fmt.Sprintf("test-cookie=%s", cookie))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, getReq)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v expected %v", status, http.StatusOK)
	}

	// Send empty data to the form
	rr = httptest.NewRecorder()
	postReq, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Error("failed to create request: POST /")
	}
	postReq.Header.Add("Cookie", fmt.Sprintf("test-cookie=%s", cookie))
	handler.ServeHTTP(rr, postReq)
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v expected %v", status, http.StatusSeeOther)
	}
}

func TestLogRoutes(t *testing.T) {
	f := new(form.Form)
	f.Pages = []form.Page{{Elements: make([]form.Element, 1)}}
	srv, err := NewService(*f, nil, echoAction, Config{CookieName: "test-cookie"})
	if err != nil {
		t.Fatalf("failed to initialise tonic service: %s", err.Error())
	}
	handler := srv.web.Handler

	// Add test cookie to the database
	testSession := db.NewSession("test-token", 42)
	srv.db.InsertSession(testSession)

	cookie := testSession.ID

	jobLabel := "TestJob"

	checkLogJobCount := func(route string, nexpected int) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", route, nil)
		if err != nil {
			t.Errorf("failed to create request: %s", route)
		}
		req.Header.Add("Cookie", fmt.Sprintf("test-cookie=%s", cookie))
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v expected %v", status, http.StatusOK)
		}

		content, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Errorf("failed to read response body: %s", err.Error())
		}
		if njobs := bytes.Count(content, []byte(jobLabel)); njobs != nexpected {
			t.Errorf("Job log returned %d, expected %d", njobs, nexpected)
		}
	}

	checkLogJobCount("/log", 0)

	srv.db.InsertJob(&db.Job{ID: 12, UserID: 42, Label: jobLabel})
	checkLogJobCount("/log", 1)

	srv.db.InsertJob(&db.Job{ID: 16, UserID: 42, Label: jobLabel})
	checkLogJobCount("/log", 2)

	srv.db.InsertJob(&db.Job{ID: 26, UserID: 44, Label: jobLabel}) // other user
	checkLogJobCount("/log", 2)
}
