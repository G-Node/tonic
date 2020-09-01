package web

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/G-Node/tonic/templates"
	"github.com/gorilla/mux"
)

// ErrorResponse logs an error and renders an error page with the given message,
// returning the given status code to the user.
func (ws *Server) ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)

	tmpl := template.New("layout")
	tmpl, err := tmpl.Parse(templates.Layout)
	if err != nil {
		tmpl = template.New("content")
	}
	tmpl, err = tmpl.Parse(templates.Fail)
	if err != nil {
		w.Write([]byte(message))
		return
	}
	errinfo := struct {
		StatusCode int
		StatusText string
		Message    string
	}{
		status,
		http.StatusText(status),
		message,
	}
	if err := tmpl.Execute(w, &errinfo); err != nil {
		log.Printf("Error rendering fail page: %v", err)
	}
}

// Server implements the web server for the Tonic service.
type Server struct {
	*http.Server
	Router *mux.Router
}

// New returns a web Server with an initialised mux.Router and http.Server.
func New(port uint16) *Server {
	srv := new(Server)
	srv.Router = new(mux.Router)
	httpsrv := new(http.Server)
	httpsrv.Handler = srv.Router

	httpsrv.Addr = fmt.Sprintf(":%d", port)
	// Good practice to set timeouts to avoid Slowloris attacks.
	httpsrv.WriteTimeout = time.Second * 15
	httpsrv.ReadTimeout = time.Second * 15
	httpsrv.IdleTimeout = time.Second * 60
	srv.Server = httpsrv
	return srv
}

// Start starts the embedded web server's ListenAndServe method in a goroutine
// and returns.  This method does not block. Use WaitForInterrupt() or
// implement your own blocking function to wait for any other stop condition.
func (ws *Server) Start() {
	go func() {
		if err := ws.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
}

// Stop gracefully stops the web service.
func (ws *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Gracefully shut down, waiting for the timeout deadline for connections to close.
	ws.Shutdown(ctx)
}
