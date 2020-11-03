package worker

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/G-Node/gin-cli/ginclient"
	ginconfig "github.com/G-Node/gin-cli/ginclient/config"
	"github.com/G-Node/gin-cli/git"
	"github.com/G-Node/tonic/tonic/db"
	"github.com/G-Node/tonic/tonic/form"
	"github.com/gogs/go-gogs-client"
)

// PreAction is a function that receives the Form struct as defined for the
// service.  It should return modified Form struct with values, constraints, or
// elements modified based on the permissions or actions supported for the bot
// and/or user, or any other external constraint that the function can
// evaluate.
type PreAction func(f form.Form, botClient, userClient *Client) (*form.Form, error)

// PostAction is a function that receives the form values when the form is
// submitted.  It should perform actions for the user through the service given
// the form values and return a list of messages and/or an error if it fails.
type PostAction func(v map[string][]string, botClient, userClient *Client) ([]string, error)

// Client embeds gogs.Client to extend functionality with new convenience
// methods.  (New clients may be added in the future using the same interface).
type Client struct {
	// Embedded GOGS API client
	*gogs.Client
	// GIN client for running git and git-annex operations. Also implements
	// some of the GOGS client functionality.
	GIN    *ginclient.Client
	webURL string
	gitURL string
	token  string
}

// NewClient returns a new worker Client.
func NewClient(webURL, gitURL, token string) *Client {
	gogsClient := gogs.NewClient(webURL, token)
	return &Client{Client: gogsClient, webURL: webURL, gitURL: gitURL, token: token}
}

// NewGINClient logs in to the GIN server, sets up the local configuration, and
// returns a new ginclient.Client instance for running git and git-annex
// commands.
func (client *Client) NewGINClient() (*ginclient.Client, error) {
	webcfg, err := ginconfig.ParseWebString(client.webURL)
	if err != nil {
		return nil, err
	}

	gitcfg, err := ginconfig.ParseGitString(client.gitURL)
	if err != nil {
		return nil, err
	}

	srvcfg := ginconfig.ServerCfg{Web: webcfg, Git: gitcfg}
	hostkeystr, _, err := git.GetHostKey(gitcfg)
	if err != nil {
		return nil, err
	}
	srvcfg.Git.HostKey = hostkeystr
	ginconfig.AddServerConf("gin", srvcfg)
	// Update known hosts file
	err = git.WriteKnownHosts()
	if err != nil {
		return nil, err
	}
	return ginclient.New("gin"), nil
}

// CloneRepo clones repository 'repo' into directory 'destdir'. The repository
// should be in the form user/repository, without any server information. The
// server address is configured in the client.
func (client *Client) CloneRepo(repo, destdir string) error {
	// NOTE: cloneRepo changes the working directory to the cloned repository
	// See: https://github.com/G-Node/gin-cli/issues/225
	// This will need to change when that issue is fixed
	origdir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(origdir)
	err = os.Chdir(destdir)
	if err != nil {
		return err
	}

	clonechan := make(chan git.RepoFileStatus)
	go client.GIN.CloneRepo(strings.ToLower(repo), clonechan)
	for stat := range clonechan {
		log.Print(stat)
		if stat.Err != nil {
			return stat.Err
		}
	}

	downloadchan := make(chan git.RepoFileStatus)
	go client.GIN.GetContent(nil, downloadchan)
	for stat := range downloadchan {
		log.Print(stat)
		if stat.Err != nil {
			return stat.Err
		}
	}
	return nil

}

// UserJob extends db.Job with a user token to perform authenticated tasks on
// behalf of a given user.
type UserJob struct {
	*db.Job
	client *Client
}

// NewUserJob returns a new UserJob initialised with the given custom function
// and user values.
func NewUserJob(client *Client, label string, values map[string][]string) *UserJob {
	j := new(UserJob)
	j.Job = new(db.Job)
	j.client = client
	user, _ := client.GetSelfInfo() // TODO: Handle error
	j.Label = label
	j.UserID = user.ID
	// copy values to avoid mutating ValueMap after it's assigned.
	j.ValueMap = make(map[string][]string, len(values))
	for k, v := range values {
		j.ValueMap[k] = v
	}
	return j
}

// Worker pool with queue for running Jobs asynchronously.
type Worker struct {
	queue chan *UserJob
	// Sending any value through 'stop' will stop the worker.
	stop chan bool
	// PreAction is used to prepare data to show the user, such as populating
	// form lists or showing information on static pages.
	PreAction PreAction
	// PostAction
	PostAction PostAction
	db         *db.Connection
	// client is used to perform administrative actions as the bot user that
	// represents the service.
	client *Client
	log    *log.Logger
}

// New returns a new Worker attached to the given database.
func New(dbconn *db.Connection) *Worker {
	w := new(Worker)
	// Set default logger.
	// Can be later replaced using the SetLogger() method.
	w.log = log.New(os.Stderr, "", log.LstdFlags)

	// TODO: Define worker queue length in configuration
	w.queue = make(chan *UserJob, 100)
	w.stop = make(chan bool)
	w.db = dbconn
	return w
}

// SetLogger sets the logger instance for the worker service.  If unset the
// service defines its own logger with the same configuration as the standard
// Logger.
func (w *Worker) SetLogger(l *log.Logger) {
	w.log = l
}

// SetClient assigns a service (bot) Client to the worker.
func (w *Worker) SetClient(c *Client) {
	w.client = c
}

// Enqueue adds the job to the queue and stores it in the database.
func (w *Worker) Enqueue(j *UserJob) {
	j.Lock()
	w.log.Printf("J: %+v", j)
	j.SubmitTime = time.Now()
	j.Unlock()
	err := w.db.InsertJob(j.Job)
	if err != nil {
		w.log.Printf("Error inserting job %+v into db: %v", j, err)
	}
	w.queue <- j
}

// PreprocessForm runs the defined PreAction and returns a modified Form.
func (w *Worker) PreprocessForm(f *form.Form, userClient *Client) (*form.Form, error) {
	if f == nil || w.PreAction == nil {
		// nothing to do
		return f, nil
	}
	botClient := w.client
	return w.PreAction(*f, botClient, userClient)
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
	j.Lock()
	defer j.Unlock()
	defer w.db.UpdateJob(j.Job) // Update job entry in db when done
	var msgs []string
	var err error
	if w.PostAction != nil {
		msgs, err = w.PostAction(j.ValueMap, w.client, j.client)
	} else {
		j.Messages = []string{}
	}
	j.Messages = msgs
	j.EndTime = time.Now()
	if err == nil {
		w.log.Printf("Job [J%d] %s finished", j.ID, j.Label)
	} else {
		w.log.Printf("Job [J%d]  %s failed: %s", j.ID, j.Label, err)
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
