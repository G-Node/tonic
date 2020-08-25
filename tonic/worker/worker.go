package worker

import (
	"log"
	"time"

	"github.com/G-Node/tonic/tonic/db"
	"github.com/gogs/go-gogs-client"
)

// JobAction is the type of the custom function that needs to be defined for
// all UserJobs.
type JobAction func(v map[string]string, botClient, userClient *Client) ([]string, error)

// Client embeds gogs.Client to extend functionality with new convenience
// methods.  (New clients may be added in the future using the same interface).
type Client struct {
	*gogs.Client
}

// NewClient returns a new worker Client.
func NewClient(url, token string) *Client {
	gc := gogs.NewClient(url, token)
	return &Client{Client: gc}
}

// UserJob extends db.Job with a user token to perform authenticated tasks on
// behalf of a given user.
type UserJob struct {
	*db.Job
	client *Client
}

// NewUserJob returns a new UserJob initialised with the given custom function
// and user values.
func NewUserJob(client *Client, values map[string]string) *UserJob {
	j := new(UserJob)
	j.Job = new(db.Job)
	j.client = client
	user, _ := client.GetSelfInfo() // TODO: Handle error
	j.UserID = user.ID
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
	client *Client
}

// New returns a new Worker attached to the given database.
func New(dbconn *db.Connection) *Worker {
	w := new(Worker)
	// TODO: Define worker queue length in configuration
	w.queue = make(chan *UserJob, 100)
	w.stop = make(chan bool)
	w.db = dbconn
	return w
}

// SetClient assigns a service (bot) Client to the worker.
func (w *Worker) SetClient(c *Client) {
	w.client = c
}

// Enqueue adds the job to the queue and stores it in the database.
func (w *Worker) Enqueue(j *UserJob) {
	log.Printf("J: %+v", j)
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

// Stop sends the stop signal to the worker pool and closes the Job channel.
func (w *Worker) Stop() {
	// TODO: Finish ongoing jobs?
	w.stop <- true
}

// run starts the custom function of the given job. When the job is
// finished, it updates it with the returned messages and error (if any) and
// updates the corresponding database entry.
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

// Start the worker queue, reading jobs sequentially from the channel and
// executing their custom function.
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
}
