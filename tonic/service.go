package tonic

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/G-Node/tonic/tonic/db"
	"github.com/G-Node/tonic/tonic/web"
	"github.com/G-Node/tonic/tonic/worker"
	"github.com/gogs/go-gogs-client"
)

// Config containing all the configuration values for a service.
type Config struct {
	GINServer   string
	Port        uint16
	CookieName  string
	DBPath      string
	GINPassword string
}

// Tonic represents a full service which contains a web server, a database for
// jobs and sessions, and a worker pool that runs the jobs.
type Tonic struct {
	web    *web.Server
	db     *db.Connection
	worker *worker.Worker
	log    *log.Logger // TODO: Move all log messages to this logger
	form   []Element
	Config *Config
}

// NewService creates a new Tonic with a given form and custom job action.
func NewService(form []Element, f worker.JobAction) (*Tonic, error) {
	srv := new(Tonic)
	// DB
	// TODO: Define db path in config
	log.Print("Initialising database")
	conn, err := db.New("./test.db")
	if err != nil {
		return nil, err
	}
	srv.db = conn

	// Worker
	log.Print("Starting worker")
	srv.worker = worker.New(srv.db)

	// Web server
	srv.web = web.New()
	srv.setupWebRoutes()

	// set form and func
	srv.SetForm(form)
	srv.SetJobAction(f)

	// TODO: Set up logger
	return srv, nil
}

// login to configured GIN server as the bot user that represents this service
// and attach a new authenticated gogs.Client to the service struct.
func (srv *Tonic) login() error {
	passfile, err := os.Open("testbot") // TODO: Add to config

	passdata, err := ioutil.ReadAll(passfile)
	if err != nil {
		return err
	}

	userpass := make(map[string]string)

	err = json.Unmarshal(passdata, &userpass)
	if err != nil {
		return err
	}

	client := gogs.NewClient(srv.Config.GINServer, "")
	tokens, err := client.ListAccessTokens(userpass["username"], userpass["password"])
	if err != nil {
		return err
	}

	var token *gogs.AccessToken
	if len(tokens) > 0 {
		token = tokens[0]
	} else {
		token, err = client.CreateAccessToken(userpass["username"], userpass["password"], gogs.CreateAccessTokenOption{Name: "testbot"})
		if err != nil {
			return err
		}
	}
	srv.worker.SetClient(worker.NewClient(srv.Config.GINServer, userpass["username"], token.Sha1))
	return nil
}

// Start the service (worker and web server).
func (srv *Tonic) Start() error {
	if srv.form == nil || len(srv.form) == 0 {
		return fmt.Errorf("nil or empty form is invalid")
	}
	if srv.worker.Action == nil {
		return fmt.Errorf("nil job function is invalid")
	}

	log.Print("Starting worker")
	srv.worker.Start()
	log.Print("Worker started")

	log.Print("Starting web service")
	srv.web.Start()
	log.Print("Web server started")

	log.Print("Logging in to gin")
	srv.login()
	log.Printf("Logged in and ready")
	return nil
}

// WaitForInterrupt blocks until the service receives an interrupt signal (SIGINT).
func (srv *Tonic) WaitForInterrupt() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
}

// Stop the service by gracefully shutting down the web service, stopping the
// worker pool, and closing the database connection, in that order.
func (srv *Tonic) Stop() {
	log.Print("Stopping web service")
	srv.web.Stop()

	log.Print("Stopping worker queue")
	srv.worker.Stop()

	log.Print("Closing database connection")
	if err := srv.db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
	log.Print("Service stopped")
}

// SetForm can be used to set or override the form for the service.
func (srv *Tonic) SetForm(form []Element) {
	srv.form = make([]Element, len(form))
	copy(srv.form, form)
}

// SetJobAction can be used to set or override the custom job action for the service.
func (srv *Tonic) SetJobAction(f worker.JobAction) {
	srv.worker.Action = f
}
