package tonic

import (
	"log"
	"os"
	"os/signal"

	"github.com/G-Node/tonic/tonic/db"
	"github.com/G-Node/tonic/tonic/web"
	"github.com/G-Node/tonic/tonic/worker"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

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
	config *Config
}

func NewService() *Tonic {
	srv := new(Tonic)
	// DB
	// TODO: Define db path in config
	log.Print("Initialising database")
	conn, err := db.New("./test.db")
	checkError(err)
	srv.db = conn

	// Worker
	log.Print("Starting worker")
	srv.worker = worker.New()

	// Web server
	srv.web = web.New()

	// TODO: Set up logger
	return srv
}

// Start the service (worker and web server).
func (srv *Tonic) Start() {
	log.Print("Starting worker")
	srv.worker.Start()
	log.Print("Starting web service")
	srv.setupWebRoutes()
	srv.web.Start()
}

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

func (t *Tonic) SetForm(form []Element) {
	t.form = make([]Element, len(form))
	copy(t.form, form)
}
