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
	router.HandleFunc("/", srv.ProcessForm).Methods("POST")
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

	client := gogs.NewClient(srv.config.GINServer, "")
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
		Name:    srv.config.CookieName,
		Value:   userToken, // create session IDs linked to token instead
		Expires: time.Now().Add(7 * 24 * time.Hour),
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

	data := make(map[string]interface{})
	elements := make([]Element, len(srv.form))
	copy(elements, srv.form)
	for _, element := range elements {
		val, ok := job.ValueMap[element.Name]
		if ok {
			element.Value = val
		}
	}
	data["elements"] = elements
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

	// TODO: Get log from database
	joblog := make([]db.JobInfo, 0)
	if err := tmpl.Execute(w, joblog); err != nil {
		log.Printf("Failed to render log: %v", err)
	}
}
