package worker

import (
	"github.com/G-Node/tonic/tonic/db"
	"log"
	"time"
)

// Worker pool with queue for running Jobs asynchronously.
type Worker struct {
	queue chan *db.Job
	stop  chan bool
}

// Job extends the db.JobInfo struct with an action.
type Job struct {
	*db.JobInfo
	Action func() error
}

func New() *Worker {
	w := new(Worker)
	// TODO: Define worker queue length in configuration
	w.queue = make(chan *db.Job, 100)
	w.stop = make(chan bool)
	return w
}

// Enqueue adds the job to the queue and stores it in the database.
func (w *Worker) Enqueue(j *db.Job) {
	w.queue <- j
	j.SubmitTime = time.Now()
}

func (w *Worker) Stop() {
	// TODO: Finish ongoing jobs?
	w.stop <- true
}

func (w *Worker) run(j *db.Job) {
	log.Printf("Starting job %q", j.Label)
	err := j.Action()
	j.EndTime = time.Now()
	if err == nil {
		log.Printf("Job [J%d] %s finished", j.ID, j.Label)
	} else {
		log.Printf("Job [J%d]  %s failed: %s", j.ID, j.Label, err)
		j.Message = err.Error()
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			select {
			case job := <-w.queue:
				run(job)
			case <-w.stop:
				return
			}
		}
	}()
	log.Print("Worker started")
}
