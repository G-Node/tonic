package worker

import (
	"log"
	"time"

	"github.com/G-Node/tonic/tonic/db"
	"github.com/gogs/go-gogs-client"
)

type JobAction func(v map[string]string, botClient, userClient *gogs.Client) ([]string, error)

// UserJob extends db.Job with a user token to perform authenticated tasks on
// behalf of a given user.
type UserJob struct {
	*db.Job
	client *gogs.Client
}

func NewUserJob(client *gogs.Client, values map[string]string) *UserJob {
	j := new(UserJob)
	j.client = client
	// copy values to avoid mutating ValueMap after it's assigned.
	j.ValueMap = make(map[string]string, len(values))
	for k, v := range values {
		j.ValueMap[k] = v
	}
	return j
}

// Worker pool with queue for running Jobs asynchronously.
type Worker struct {
	queue  chan *UserJob
	stop   chan bool
	Action JobAction
	db     *db.Connection
	// client is used to perform administrative actions as the bot user that
	// represents the srevice.
	client *gogs.Client
}

func New(dbconn *db.Connection) *Worker {
	w := new(Worker)
	// TODO: Define worker queue length in configuration
	w.queue = make(chan *UserJob, 100)
	w.stop = make(chan bool)
	w.db = dbconn
	return w
}

func (w *Worker) SetClient(c *gogs.Client) {
	w.client = c
}

// Enqueue adds the job to the queue and stores it in the database.
func (w *Worker) Enqueue(j *UserJob) {
	j.SubmitTime = time.Now()
	// TODO: Find a good way to label jobs otherwise just use IDs in listings
	var label string
	for _, label = range j.ValueMap {
		break
	}
	j.Label = label
	err := w.db.InsertJob(j.Job)
	if err != nil {
		log.Printf("Error inserting job %+v into db: %v", j, err)
	}
	w.queue <- j
}

func (w *Worker) Stop() {
	// TODO: Finish ongoing jobs?
	w.stop <- true
}

func (w *Worker) run(j *UserJob) {
	defer w.db.UpdateJob(j.Job) // Update job entry in db when done
	msgs, err := w.Action(j.ValueMap, w.client, j.client)
	j.EndTime = time.Now()
	j.Messages = msgs
	if err == nil {
		log.Printf("Job [J%d] %s finished", j.ID, j.Label)
	} else {
		log.Printf("Job [J%d]  %s failed: %s", j.ID, j.Label, err)
		j.Error = err.Error()
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
