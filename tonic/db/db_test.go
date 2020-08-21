package db

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestInitEmpty(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "testdb")
	if err != nil {
		t.Fatalf("Failed to create temporary database file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	db, err := New(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to initialise database connection to file %q: %s", tmpfile.Name(), err.Error())
	}
	defer db.Close()

	// db should be empty
	jobs, err := db.AllJobs()
	if err != nil {
		t.Fatalf("Failed to retrieve all jobs from empty db: %s", err.Error())
	}

	if jobs == nil {
		t.Fatal("Job listing returned nil instead of empty slice")
	}

	if len(jobs) != 0 {
		t.Fatalf("Job listing returned %d entries; should be 0", len(jobs))
	}

	sessions := make([]Session, 0)
	if err := db.engine.Find(&sessions); err != nil {
		t.Fatalf("Failed to retrieve all sessions from empty db: %s", err.Error())
	}

	if len(sessions) != 0 {
		t.Fatalf("Session listing returned %d entries; should be 0", len(sessions))
	}
}

func TestSessionStore(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "testdb")
	if err != nil {
		t.Fatalf("Failed to create temporary database file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	db, err := New(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to initialise database connection to file %q: %s", tmpfile.Name(), err.Error())
	}
	defer db.Close()

	empty := &Session{}
	if db.InsertSession(empty) != nil {
		t.Fatalf("Failed inserting empty session: %s", err.Error())
	}

	if db.InsertSession(empty) == nil {
		t.Fatal("Succeeded while entering duplicate empty session")
	}

	if s, err := db.GetSession(""); err != nil {
		t.Fatalf("Failed to retrieve empty session: %s", err.Error())
	} else if s.ID != "" {
		t.Fatalf("Unexpected session returned from db: %+v", s)
	}

	if err := db.DeleteSession(empty.ID); err != nil {
		t.Fatalf("Failed to delete empty session: %s", err.Error())
	}

	sess := NewSession("testuser", "faketoken")
	if db.InsertSession(sess) != nil {
		t.Fatalf("Failed inserting new session: %s", err.Error())
	}

	if db.InsertSession(sess) == nil {
		t.Fatal("Succeeded inserting duplicate session")
	}

	dupe := NewSession("otheruser", "anothertoken")
	dupe.ID = sess.ID
	if db.InsertSession(dupe) == nil {
		t.Fatal("Succeeded inserting session with conflicting ID")
	}

	sessions := make([]Session, 0)
	if err := db.engine.Find(&sessions); err != nil {
		t.Fatalf("Failed to retrieve all sessions from db: %s", err.Error())
	}

	if s, err := db.GetSession(sess.ID); err != nil {
		t.Fatalf("Failed to retrieve test session from db: %s", err.Error())
	} else if s.ID != sess.ID {
		t.Fatalf("Unexpected session returned from db: %+v (not %+v)", s, sess)
	}

	for _, s := range sessions {
		db.DeleteSession(s.ID)
	}

	sessions = make([]Session, 0)
	if err := db.engine.Find(&sessions); err != nil {
		t.Fatalf("Failed to retrieve all sessions from db: %s", err.Error())
	}
	if len(sessions) > 0 {
		t.Fatalf("Unexpected sessions found in db after deletion: %+v", sessions)
	}
}
