package tonic

import (
	"fmt"
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
	testSession := db.NewSession("test-token")
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
