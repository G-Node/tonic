package worker

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/G-Node/tonic/tonic/db"
)

func TestWorkerEmptyJob(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "testdb")
	if err != nil {
		t.Fatalf("Failed to create temporary database file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	conn, err := db.New(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to initialise database connection to file %q: %s", tmpfile.Name(), err.Error())
	}
	defer conn.Close()

	w := New(conn)
	w.PostAction = testAction
	w.Start()
	defer w.Stop()
	j := new(UserJob)
	j.client = new(Client)
	j.Job = new(db.Job)
	w.Enqueue(j)
	time.Sleep(time.Millisecond)
	if !j.IsFinished() {
		t.Fatalf("Job not finished: %+v", j)
	}
	if len(j.Messages) > 0 {
		t.Fatalf("Job has messages when it shouldn't: %v", j.Messages)
	}
}

func TestWorkerAction(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "testdb")
	if err != nil {
		t.Fatalf("Failed to create temporary database file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	conn, err := db.New(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to initialise database connection to file %q: %s", tmpfile.Name(), err.Error())
	}
	defer conn.Close()

	w := New(conn)
	w.PostAction = testAction
	w.Start()
	defer w.Stop()
	w.client = NewClient("https://example.org", "testadmintoken")
	j := NewUserJob(NewClient("https://example.org", "testusertoken"), "workertest", map[string][]string{"A": {"alpha"}, "Z": {"zeta"}})
	w.Enqueue(j)
	time.Sleep(time.Millisecond)
	if !j.IsFinished() {
		t.Fatalf("Job not finished: %+v", j)
	}
	if len(j.Messages) != 2 {
		t.Fatalf("Unexpected number of messages found in finished job: %d != 2", len(j.Messages))
	}
	if j.Error != "" {
		t.Fatalf("Job failed with error: %s", j.Error)
	}
}

func TestWorkerJobFail(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "testdb")
	if err != nil {
		t.Fatalf("Failed to create temporary database file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	conn, err := db.New(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to initialise database connection to file %q: %s", tmpfile.Name(), err.Error())
	}
	defer conn.Close()

	w := New(conn)
	w.PostAction = testAction
	w.Start()
	defer w.Stop()
	w.client = NewClient("https://example.org", "testadmintoken")
	j := NewUserJob(NewClient("https://example.org", "testusertoken"), "workerjobfailtest", map[string][]string{"A": {"error"}, "Ω": {"omega"}})
	w.Enqueue(j)
	time.Sleep(time.Millisecond)
	if !j.IsFinished() {
		t.Fatalf("Job not finished: %+v", j)
	}
	if len(j.Messages) != 2 {
		t.Fatalf("Unexpected number of messages found in finished job: %d != 2", len(j.Messages))
	}
	if j.Error == "" {
		t.Fatal("Job succeeded when it should have failed")
	}
}

func testAction(values map[string][]string, bc, uc *Client) ([]string, error) {
	// Simply return each key:value pair as separate lines in messages
	// If any value is the string 'error', return with error.
	m := make([]string, 0, len(values)+2)
	log.Printf("Got %d values", len(values))
	var err error
	for k, v := range values {
		m = append(m, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))
		if len(v) > 0 && v[0] == "error" {
			err = fmt.Errorf("Found error value in key %q", k)
		}
	}

	return m, err
}
