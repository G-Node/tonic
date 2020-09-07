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

	"github.com/G-Node/tonic/tonic/worker"
)

func TestTonicFailStart(t *testing.T) {
	if s, _ := NewService(nil, nil, Config{}); s.Start() == nil {
		s.Stop()
		t.Fatal("Service start succeeded; should have failed")
	}

	if s, _ := NewService(make([]Element, 0), nil, Config{}); s.Start() == nil {
		s.Stop()
		t.Fatal("Service start succeeded; should have failed")
	}

	if s, _ := NewService(make([]Element, 10), nil, Config{}); s.Start() == nil {
		s.Stop()
		t.Fatal("Service start succeeded; should have failed")
	}
}

func TestTonicWithForm(t *testing.T) {
	f := []Element{
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
	srv, err := NewService(f, noopAction, Config{})
	if err != nil {
		t.Fatalf("Failed to initialise tonic service: %s", err.Error())
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start tonic service: %s", err.Error())
	}

	srv.Stop()
}

func noopAction(values map[string]string, _, _ *worker.Client) ([]string, error) {
	return nil, nil
}

func TestTonicWithAction(t *testing.T) {
	srv, err := NewService(make([]Element, 1), echoAction, Config{})
	if err != nil {
		t.Fatalf("Failed to initialise tonic service: %s", err.Error())
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start tonic service: %s", err.Error())
	}

	j := worker.NewUserJob(worker.NewClient("", ""), map[string]string{"α": "alpha", "ω": "omega"})
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

func echoAction(values map[string]string, _, _ *worker.Client) ([]string, error) {
	echo := make([]string, 0, len(values))
	for k, v := range values {
		echo = append(echo, fmt.Sprintf("%s:%s", k, v))
	}

	sort.Strings(echo)
	return echo, nil
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
	srv, err := NewService(make([]Element, 1), noopAction, Config{})
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
