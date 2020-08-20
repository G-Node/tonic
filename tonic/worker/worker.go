package worker

import (
	"log"
	"time"

	"github.com/G-Node/tonic/tonic/db"
)

type JobAction func(v map[string]string) ([]string, error)

// Worker pool with queue for running Jobs asynchronously.
type Worker struct {
	queue  chan *db.Job
	stop   chan bool
	Action JobAction
	db     *db.Connection
}

func New(dbconn *db.Connection) *Worker {
	w := new(Worker)
	// TODO: Define worker queue length in configuration
	w.queue = make(chan *db.Job, 100)
	w.stop = make(chan bool)
	w.db = dbconn
	return w
}

// Enqueue adds the job to the queue and stores it in the database.
func (w *Worker) Enqueue(j *db.Job) {
	j.SubmitTime = time.Now()
	var label string
	for _, label = range j.ValueMap {
		break
	}
	j.Label = label
	err := w.db.InsertJob(j)
	if err != nil {
		log.Printf("Error inserting job %+v into db: %v", j, err)
	}
	w.queue <- j
}

func (w *Worker) Stop() {
	// TODO: Finish ongoing jobs?
	w.stop <- true
}

func (w *Worker) run(j *db.Job) {
	defer w.db.UpdateJob(j) // Update job entry in db when done
	log.Printf("Starting job %q", j.Label)
	err := w.JobFunc(j.ValueMap)
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
				w.run(job)
			case <-w.stop:
				return
			}
		}
	}()
	log.Print("Worker started")
}
