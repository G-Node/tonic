package db

import (
	"fmt"
	"time"
)

// JobInfo holds all the information for a given Job.
type JobInfo struct {
	// Job ID (auto)
	ID int64 `xorm:"pk autoincr"`
	// ID of user who submitted the job
	UserID int64
	// Name/label of the job
	Label string
	// Message returned from finished job
	Message string
	// Form values that created the job
	ValueMap map[string]string
	// Time when the job was submitted to the queue
	SubmitTime time.Time
	// Time when the job finished (0 if ongoing)
	EndTime time.Time
}

// InsertJob inserts a new Job into the database.  Upon successful return, the
// Job has a new unique ID.
func (conn *Connection) InsertJob(job *JobInfo) error {
	_, err := conn.engine.Insert(job) // job ID is assigned on insertion
	return err
}

// UpdateJob updates an existing Job entry in the database.
func (conn *Connection) UpdateJob(job *JobInfo) error {
	_, err := conn.engine.Update(job)
	return err
}

// GetUserJobs retrieves all the Jobs associated with a given UserID.
func (conn *Connection) GetUserJobs(uid int64) ([]JobInfo, error) {
	var userjobs []JobInfo
	condition := JobInfo{UserID: uid}
	if err := conn.engine.Find(&userjobs, condition); err != nil {
		return nil, err
	}

	return userjobs, nil
}

// IsFinished returns true if the Job has finished (has an EndTime).
func (ji *JobInfo) IsFinished() bool {
	return !ji.EndTime.IsZero()
}

// AllJobs returns all Job entries in the database.
func (conn *Connection) AllJobs() ([]JobInfo, error) {
	var alljobs []JobInfo
	if err := conn.engine.Find(&alljobs); err != nil {
		return nil, err
	}

	return alljobs, nil
}

// GetJob retrieves a Job from the database given its ID.
func (conn *Connection) GetJob(id int64) (*JobInfo, error) {
	ji := new(JobInfo)
	ji.ID = id
	if has, err := conn.engine.Get(ji); err != nil {
		return nil, err
	} else if !has {
		return nil, fmt.Errorf("not found")
	}
	return ji, nil
}
