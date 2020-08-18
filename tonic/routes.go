// Common routes and pages
package tonic

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/G-Node/tonic/templates"
	"github.com/G-Node/tonic/tonic/db"
	"github.com/gogs/go-gogs-client"
	"github.com/gorilla/mux"
)

// setupWebRoutes sets up the common routes shared by all instances of the service.
//
// Login, Form (editable and read-only), and Job log pages
func (srv *Tonic) setupWebRoutes() error {
	router := srv.web.Router
	router.StrictSlash(true)

	router.HandleFunc("/login", srv.renderLoginPage).Methods("GET")
	router.HandleFunc("/login", srv.userLoginPost).Methods("POST")

	router.HandleFunc("/", srv.renderForm).Methods("GET")
	router.HandleFunc("/", srv.processForm).Methods("POST")
	router.HandleFunc("/log", srv.renderLog).Methods("GET")
	router.HandleFunc("/log/{id:[0-9]+}", srv.showJob).Methods("GET")

	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	return nil
}

func (srv *Tonic) renderLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("layout")
	tmpl, err := tmpl.Parse(templates.Layout)
	checkError(err)
	tmpl, err = tmpl.Parse(templates.Login)
	checkError(err)
	tmpl.Execute(w, nil)
}

func (srv *Tonic) userLoginPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		srv.web.ErrorResponse(w, http.StatusUnauthorized, "authentication failed")
		return
	}

	client := gogs.NewClient(srv.Config.GINServer, "")
	var userToken string
	tokens, err := client.ListAccessTokens(username, password)
	if err != nil {
		srv.web.ErrorResponse(w, http.StatusUnauthorized, "authentication failed")
		return
	}

	if len(tokens) == 0 {
		token, err := client.CreateAccessToken(username, password, gogs.CreateAccessTokenOption{Name: "testbot"})
		userToken = token.Sha1
		if err != nil {
			srv.web.ErrorResponse(w, http.StatusUnauthorized, "authentication failed")
			return
		}
	} else {
		userToken = tokens[0].Sha1
	}

	// TODO: Session cookie with token in DB
	cookie := http.Cookie{
		Name:    srv.Config.CookieName,
		Value:   userToken,                          // TODO: create session IDs linked to token instead
		Expires: time.Now().Add(7 * 24 * time.Hour), // TODO: Configurable expiration
		Secure:  false,
	}

	http.SetCookie(w, &cookie)
	// Redirect to form
	http.Redirect(w, r, "/", http.StatusFound)
}

func (srv *Tonic) renderForm(w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("layout")
	tmpl, err := tmpl.Parse(templates.Layout)
	checkError(err)
	tmpl, err = tmpl.Parse(templates.Form)

	checkError(err)

	elements := make([]Element, len(srv.form))
	copy(elements, srv.form)

	data := make(map[string]interface{})
	data["elements"] = elements

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Failed to render form: %v", err)
	}
}

func (srv *Tonic) showJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobid, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Invalid ID")
		return
	}
	job, err := srv.db.GetJob(jobid)

	if err != nil || job == nil {
		srv.web.ErrorResponse(w, http.StatusNotFound, "No such job")
		return
	}

	tmpl := template.New("layout")
	tmpl, err = tmpl.Parse(templates.Layout)
	checkError(err)
	tmpl, err = tmpl.Parse(templates.Form)
	checkError(err)

	// Set up form and assign values to each matching element
	data := make(map[string]interface{})
	elements := make([]Element, len(srv.form))
	copy(elements, srv.form)
	for idx := range elements {
		if val, ok := job.ValueMap[elements[idx].Name]; ok {
			elements[idx].Value = val
		}
	}

	// Add timestamps and exit message to template data and set read-only
	data["elements"] = elements
	data["submit_time"] = job.SubmitTime
	data["end_time"] = job.EndTime
	data["message"] = job.Message
	data["readonly"] = true

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Failed to render form: %v", err)
	}
}
func (srv *Tonic) renderLog(w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("layout")
	tmpl, err := tmpl.Parse(templates.Layout)
	checkError(err)
	tmpl, err = tmpl.Parse(templates.LogView)
	checkError(err)

	joblog, err := srv.db.AllJobs()
	if err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Error reading jobs from DB")
		return
	}
	if err := tmpl.Execute(w, joblog); err != nil {
		log.Printf("Failed to render log: %v", err)
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Error showing job listing")
		return
	}
}

func (srv *Tonic) processForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("Failed to parse form: %v", err)
	}
	postValues := r.PostForm
	jobValues := make(map[string]string)
	for idx := range srv.form {
		key := srv.form[idx].Name
		jobValues[key] = postValues.Get(key)
	}

	newJob := new(db.Job)
	newJob.ValueMap = jobValues

	srv.worker.Enqueue(newJob)
}
