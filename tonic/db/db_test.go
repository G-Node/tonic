package db

import (
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
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

	sess := NewSession("faketoken")
	if db.InsertSession(sess) != nil {
		t.Fatalf("Failed inserting new session: %s", err.Error())
	}

	if db.InsertSession(sess) == nil {
		t.Fatal("Succeeded inserting duplicate session")
	}

	dupe := NewSession("anothertoken")
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

func TestJobStore(t *testing.T) {
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

	empty := &Job{}
	if db.InsertJob(empty) != nil {
		t.Fatalf("Failed inserting empty job: %s", err.Error())
	}

	if empty.ID != 1 {
		t.Fatalf("Job ID autoincrement failed: %d", empty.ID)
	}

	if db.InsertJob(empty) == nil {
		t.Fatal("Succeeded while entering duplicate empty job")
	}

	job := new(Job)
	job.Label = "test"
	job.ValueMap = map[string]string{
		"key1":       "value1",
		"key2":       "value2",
		"anotherkey": "anothervalue",
		"onemore":    "lastvalue",
	}
	if db.InsertJob(job) != nil {
		t.Fatalf("Failed inserting new job: %s", err.Error())
	}

	if db.InsertJob(job) == nil {
		t.Fatal("Succeeded inserting duplicate job")
	}

	dupe := new(Job)
	dupe.ID = job.ID
	if db.InsertJob(dupe) == nil {
		t.Fatal("Succeeded inserting job with conflicting ID")
	}

	nExpected := 2
	if jobs, err := db.AllJobs(); err != nil {
		t.Fatalf("Failed to retrieve all jobs from db: %s", err.Error())
	} else if len(jobs) != nExpected {
		t.Fatalf("Unexpected number of jobs found: %d (expected %d)", len(jobs), nExpected)
	}

	if j, err := db.GetJob(job.ID); err != nil {
		t.Fatalf("Failed to retrieve test job from db: %s", err.Error())
	} else if j.ID != job.ID {
		t.Fatalf("Unexpected job returned from db: %+v (not %+v)", j, job)
	} else if len(j.ValueMap) != len(job.ValueMap) {
		t.Fatalf("Job ValueMap mismatch: %+v (not %+v)", j, job)
	} else {
		for k := range job.ValueMap {
			if j.ValueMap[k] != job.ValueMap[k] {
				t.Fatalf("Job ValueMap value mismatch: %s (not %s)", j.ValueMap[k], job.ValueMap[k])
			}
		}
	}

	if j, err := db.GetJob(1000); err == nil {
		t.Fatalf("Succeeded retrieving job using invalid ID: %+v", j)
	}

	fjob := new(Job)
	if db.InsertJob(fjob) != nil {
		t.Fatalf("Failed to insert job in db: %v", fjob)
	}
	if fjob.IsFinished() {
		t.Fatalf("New (unfinished) job appears finished: %+v", fjob)
	}
	fjob.EndTime = time.Now()
	if !fjob.IsFinished() {
		t.Fatalf("Finished job appears unfinished: %+v", fjob)
	}
	if err := db.UpdateJob(fjob); err != nil {
		t.Fatalf("Failed to update job (finished): %s", err.Error())
	}
	if fjobr, err := db.GetJob(fjob.ID); err != nil {
		t.Fatalf("Failed to retrieve finished job from db: %s", err.Error())
	} else if !fjobr.IsFinished() {
		t.Fatalf("Finished job, loaded from db, appears unfinished: %s", err.Error())
	}
}

func TestUserJobs(t *testing.T) {
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

	// 200 entries for user 42
	testid := int64(42)
	ntest := 200
	testlabel := "testuserjob"
	for idx := 0; idx < ntest; idx++ {
		db.InsertJob(&Job{UserID: testid, SubmitTime: time.Now().Add(-time.Duration(time.Second)), EndTime: time.Now(), Label: testlabel})
	}

	// 1000 entries for random other users (not 42)
	rand.Seed(time.Now().UnixNano())
	for idx := 0; idx < 1000; idx++ {
		db.InsertJob(&Job{UserID: testid + 10 + rand.Int63(), SubmitTime: time.Now().Add(-time.Duration(time.Second)), EndTime: time.Now(), Label: "OtherJob"})
	}

	if ftjobs, err := db.GetUserJobs(testid); err != nil {
		t.Fatalf("Failed to get jobs for user %d: %s", testid, err.Error())
	} else if len(ftjobs) != ntest {
		t.Fatalf("Unexpected job count: %d (expected %d)", len(ftjobs), ntest)
	} else {
		for idx := range ftjobs {
			if ftjobs[idx].Label != testlabel {
				t.Fatalf("Unexpected label found for job: %s (expected %s)", ftjobs[idx].Label, testlabel)
			}
		}
	}

	if alljobs, err := db.AllJobs(); err != nil {
		t.Fatalf("Failed to retrieve all jobs: %s", err.Error())
	} else {
		nother := 0
		for idx := range alljobs {
			j := alljobs[idx]
			if j.UserID == testid {
				if j.Label != testlabel {
					t.Fatalf("Unexpected row found in db: UID %d; Label: %s", j.UserID, j.Label)
				}
			} else if j.Label != "OtherJob" {
				t.Fatalf("Unexpected row found in db: UID %d; Label: %s", j.UserID, j.Label)
			} else {
				nother++
			}
		}
		if nother != 1000 {
			t.Fatalf("Unexpected job count: %d (expected 1000)", nother)
		}
	}
}
