package tonic

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/G-Node/tonic/tonic/form"
	"github.com/G-Node/tonic/tonic/worker"
)

func TestTonicFailStart(t *testing.T) {
	if s, _ := NewService(form.Form{}, nil, nil, Config{}); s.Start() == nil {
		s.Stop()
		t.Fatal("Service start succeeded; should have failed")
	}
}

func TestTonicWithForm(t *testing.T) {
	elems := []form.Element{
		{
			ID:          "el1",
			Name:        "testfield1",
			Label:       "TestField1",
			Description: "Field of tests",
		},
		{
			ID:          "el2",
			Name:        "testfield2",
			Label:       "TestField2",
			Description: "Field of tests, part 2",
		},
	}
	f := new(form.Form)
	f.Pages = []form.Page{{Elements: elems}}
	srv, err := NewService(*f, nil, noopAction, Config{})
	if err != nil {
		t.Fatalf("Failed to initialise tonic service: %s", err.Error())
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start tonic service: %s", err.Error())
	}

	srv.Stop()
}

func noopAction(values map[string][]string, _, _ *worker.Client) ([]string, error) {
	return nil, nil
}

func TestTonicWithPreAction(t *testing.T) {
	f := new(form.Form)
	f.Pages = []form.Page{{Elements: make([]form.Element, 1)}}
	srv, err := NewService(*f, addElementAction, nil, Config{})
	if err != nil {
		t.Fatalf("Failed to initialise tonic service: %s", err.Error())
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start tonic service: %s", err.Error())
	}

	j := worker.NewUserJob(worker.NewClient("", "", ""), "testjob", map[string][]string{"α": {"alpha"}, "ω": {"omega"}})
	srv.worker.Enqueue(j)

	for !j.IsFinished() { // wait for job to finish
		time.Sleep(time.Millisecond)
	}

	if len(j.Messages) > 0 {
		t.Fatalf("Unexpected job output messages: %+v", j.Messages)
	}

	srv.Stop()
}

func addElementAction(f form.Form, _, _ *worker.Client) (*form.Form, error) {
	fnew := new(form.Form)
	fnew.Pages = []form.Page{{Elements: []form.Element{{Name: "Test", ID: "test", Description: "A test", Label: "Test"}}}}
	return fnew, nil
}

func TestTonicWithPostAction(t *testing.T) {
	f := new(form.Form)
	f.Pages = []form.Page{{Elements: make([]form.Element, 1)}}
	srv, err := NewService(*f, nil, echoAction, Config{})
	if err != nil {
		t.Fatalf("Failed to initialise tonic service: %s", err.Error())
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start tonic service: %s", err.Error())
	}

	j := worker.NewUserJob(worker.NewClient("", "", ""), "testjob", map[string][]string{"α": {"alpha"}, "ω": {"omega"}})
	srv.worker.Enqueue(j)

	for !j.IsFinished() { // wait for job to finish
		time.Sleep(time.Millisecond)
	}

	if j.Messages[0] != "α:alpha" {
		t.Fatalf("Unexpected job output message [0]: %q", j.Messages[0])
	}
	if j.Messages[1] != "ω:omega" {
		t.Fatalf("Unexpected job output message [1]: %s", j.Messages[1])
	}

	srv.Stop()
}

func echoAction(values map[string][]string, _, _ *worker.Client) ([]string, error) {
	echo := make([]string, 0, len(values))
	for k, v := range values {
		echo = append(echo, fmt.Sprintf("%s:%s", k, strings.Join(v, ", ")))
	}

	sort.Strings(echo)
	return echo, nil
}

func TestTonicWithActions(t *testing.T) {
	f := new(form.Form)
	f.Pages = []form.Page{{Elements: make([]form.Element, 1)}}
	srv, err := NewService(*f, addElementAction, echoAction, Config{})
	if err != nil {
		t.Fatalf("Failed to initialise tonic service: %s", err.Error())
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start tonic service: %s", err.Error())
	}

	j := worker.NewUserJob(worker.NewClient("", "", ""), "testtonicwithactions", map[string][]string{"α": {"alpha", "a"}, "ω": {"omega"}})
	srv.worker.Enqueue(j)

	for !j.IsFinished() { // wait for job to finish
		time.Sleep(time.Millisecond)
	}

	if j.Messages[0] != "α:alpha, a" {
		t.Fatalf("Unexpected job output message [0]: %q", j.Messages[0])
	}
	if j.Messages[1] != "ω:omega" {
		t.Fatalf("Unexpected job output message [1]: %s", j.Messages[1])
	}

	srv.Stop()
}

type LogBuffer struct {
	b   bytes.Buffer
	mux sync.Mutex
}

func (lb *LogBuffer) Write(b []byte) (int, error) {
	lb.mux.Lock()
	defer lb.mux.Unlock()
	return lb.b.Write(b)
}

func (lb *LogBuffer) String() string {
	lb.mux.Lock()
	defer lb.mux.Unlock()
	return lb.b.String()
}

func TestLoggers(t *testing.T) {
	f := new(form.Form)
	f.Pages = []form.Page{{Elements: make([]form.Element, 1)}}
	srv, err := NewService(*f, nil, noopAction, Config{})
	if err != nil {
		t.Fatalf("Failed to initialise tonic service: %s", err.Error())
	}

	prefix := "[tonictest] "

	lb := new(LogBuffer)
	logger := log.New(lb, prefix, 0)

	srv.SetLogger(logger)

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start tonic service: %s", err.Error())
	}
	srv.Stop()

	logstring := lb.String()

	expMessages := []string{
		"Starting worker",
		"Worker started",
		"Starting web service",
		"Web server started",
		"Stopping web service",
		"Stopping worker queue",
		"Closing database connection",
		"Service stopped",
	}

	for _, msg := range expMessages {
		expmsg := fmt.Sprintf("%s%s", prefix, msg)
		if !strings.Contains(logstring, expmsg) {
			log.Fatalf("Expected message %q not found in log", expmsg)
		}
	}
}
