package tonic

import (
	"fmt"
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
	GINUsername string
	GINPassword string
}

// Tonic represents a full service which contains a web server, a database for
// jobs and sessions, and a worker pool that runs the jobs.
type Tonic struct {
	web    *web.Server
	db     *db.Connection
	worker *worker.Worker
	log    *log.Logger
	form   *Form
	config *Config
}

// NewService creates a new Tonic with a given form and custom job action.
func NewService(form Form, f worker.JobAction, config Config) (*Tonic, error) {
	srv := new(Tonic)

	// Logger
	srv.log = log.New(os.Stderr, "tonic: ", log.LstdFlags) // TODO: Support naming the service and use the name as a logger prefix

	srv.config = &config
	// DB
	srv.log.Print("Initialising database")
	conn, err := db.New(config.DBPath)
	if err != nil {
		return nil, err
	}
	srv.db = conn

	// Worker
	srv.log.Print("Initialising worker")
	srv.worker = worker.New(srv.db)
	// Share logger with worker
	srv.worker.SetLogger(srv.log)

	// Web server
	srv.log.Print("Initialising web service")
	srv.web = web.New(config.Port)
	// Share logger with web service
	srv.web.SetLogger(srv.log)

	srv.log.Print("Setting up router")
	srv.setupWebRoutes()

	// set form and func
	srv.SetForm(form)
	srv.SetJobAction(f)

	return srv, nil
}

// SetLogger sets the logger instance for the tonic service and the included
// worker queue and web service.  If unset the service defines its own logger
// with the same configuration as the standard Logger and the prefix 'tonic: '.
func (srv *Tonic) SetLogger(l *log.Logger) {
	srv.log = l
	srv.worker.SetLogger(l)
	srv.web.SetLogger(l)
}

// login to configured GIN server as the bot user that represents this service
// and attach a new authenticated gogs.Client to the service struct.
func (srv *Tonic) login() error {
	username := srv.config.GINUsername
	password := srv.config.GINPassword

	client := gogs.NewClient(srv.config.GINServer, "")
	tokens, err := client.ListAccessTokens(username, password)
	if err != nil {
		return err
	}

	var token *gogs.AccessToken
	if len(tokens) > 0 {
		token = tokens[0]
	} else {
		token, err = client.CreateAccessToken(username, password, gogs.CreateAccessTokenOption{Name: "tonic"})
		if err != nil {
			return err
		}
	}
	srv.worker.SetClient(worker.NewClient(srv.config.GINServer, token.Sha1))
	return nil
}

// Start the service (worker and web server).
func (srv *Tonic) Start() error {
	if srv.form == nil || len(srv.form.Pages) == 0 {
		return fmt.Errorf("nil or empty form is invalid")
	}
	if srv.worker.Action == nil {
		return fmt.Errorf("nil job function is invalid")
	}

	srv.log.Print("Starting worker")
	srv.worker.Start()
	srv.log.Print("Worker started")

	srv.log.Print("Starting web service")
	srv.web.Start()
	srv.log.Print("Web server started")

	if srv.config.GINServer != "" {
		srv.log.Print("Logging in to gin")
		if err := srv.login(); err != nil {
			return err
		}
		srv.log.Print("Logged in and ready")
	}
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
	srv.log.Print("Stopping web service")
	srv.web.Stop()

	srv.log.Print("Stopping worker queue")
	srv.worker.Stop()

	srv.log.Print("Closing database connection")
	if err := srv.db.Close(); err != nil {
		srv.log.Printf("Error closing database: %v", err)
	}
	srv.log.Print("Service stopped")
}

// SetForm can be used to set or override the form for the service.
func (srv *Tonic) SetForm(form Form) {
	srv.form = new(Form)
	// copy elements manually
	srv.form.Name = form.Name
	srv.form.Description = form.Description
	srv.form.Pages = make([]Page, len(form.Pages))
	for pageIdx := range form.Pages {
		elements := make([]Element, len(form.Pages[pageIdx].Elements))
		copy(elements, form.Pages[pageIdx].Elements)
		srv.form.Pages[pageIdx].Elements = elements
		srv.form.Pages[pageIdx].Description = form.Pages[pageIdx].Description
	}
}

// SetJobAction can be used to set or override the custom job action for the service.
func (srv *Tonic) SetJobAction(f worker.JobAction) {
	srv.worker.Action = f
}
