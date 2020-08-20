package db

import (
	"fmt"
	"log"
	"time"
)

// Job holds all the information for a given Job.
type Job struct {
	// Job ID (auto)
	ID int64 `xorm:"pk autoincr"`
	// ID of user who submitted the job
	UserID int64
	// Name/label of the job
	Label string
	// Messages returned from finished job. This will be visible to the user
	// once the job is finished.
	Messages []string
	// Error message (if the job failed).
	Error string
	// Form values that created the job
	ValueMap map[string]string
	// Time when the job was submitted to the queue
	SubmitTime time.Time
	// Time when the job finished (0 if ongoing)
	EndTime time.Time
}

// InsertJob inserts a new Job into the database.  Upon successful return, the
// Job has a new unique ID.
func (conn *Connection) InsertJob(job *Job) error {
	_, err := conn.engine.Insert(job) // job ID is assigned on insertion
	return err
}

// UpdateJob updates an existing Job entry in the database.
func (conn *Connection) UpdateJob(job *Job) error {
	// Update only job matching the same ID
	_, err := conn.engine.Update(job, &Job{ID: job.ID})
	if err != nil {
		log.Printf("Failed to update job %d in DB: %s", job.ID, err.Error())
	}
	return err
}

// GetUserJobs retrieves all the Jobs associated with a given UserID.
func (conn *Connection) GetUserJobs(uid int64) ([]Job, error) {
	var userjobs []Job
	condition := Job{UserID: uid}
	if err := conn.engine.Find(&userjobs, condition); err != nil {
		return nil, err
	}

	return userjobs, nil
}

// IsFinished returns true if the Job has finished (has an EndTime).
func (j *Job) IsFinished() bool {
	return !j.EndTime.IsZero()
}

// AllJobs returns all Job entries in the database.
func (conn *Connection) AllJobs() ([]Job, error) {
	var alljobs []Job
	if err := conn.engine.Find(&alljobs); err != nil {
		return nil, err
	}

	return alljobs, nil
}

// GetJob retrieves a Job from the database given its ID.
func (conn *Connection) GetJob(id int64) (*Job, error) {
	j := new(Job)
	j.ID = id
	if has, err := conn.engine.Get(j); err != nil {
		return nil, err
	} else if !has {
		return nil, fmt.Errorf("not found")
	}
	return j, nil
}
