package tonic

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/G-Node/tonic/templates"
	"github.com/G-Node/tonic/tonic/db"
	"github.com/G-Node/tonic/tonic/worker"
	"github.com/gogs/go-gogs-client"
	"github.com/gorilla/mux"
)

// authedHandler is a handler that requires an authenticated user
type authedHandler func(w http.ResponseWriter, r *http.Request, session *db.Session)

// reqLoginHandler acts as middleware to check if the user is logged in.
// Returns a function that matches 'authedHandler()'.
// Use for pages that require authentication (currently, everything except the login page).
func (srv *Tonic) reqLoginHandler(handler authedHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(srv.config.CookieName)
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		sessid := cookie.Value
		session, err := srv.db.GetSession(sessid)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// TODO: Check that the session is still valid (by checking expiration)
		handler(w, r, session)
	}
}

// setupWebRoutes sets up the common routes shared by all instances of the service.
//
// Login, Form (editable and read-only), and Job log pages
func (srv *Tonic) setupWebRoutes() error {
	router := srv.web.Router
	router.StrictSlash(true)

	router.HandleFunc("/login", srv.renderLoginPage).Methods("GET")
	router.HandleFunc("/login", srv.userLoginPost).Methods("POST")

	router.HandleFunc("/", srv.reqLoginHandler(srv.renderForm)).Methods("GET")
	router.HandleFunc("/", srv.reqLoginHandler(srv.processForm)).Methods("POST")
	router.HandleFunc("/log", srv.reqLoginHandler(srv.renderLog)).Methods("GET")
	router.HandleFunc("/log/{id:[0-9]+}", srv.reqLoginHandler(srv.showJob)).Methods("GET")

	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	return nil
}

func (srv *Tonic) renderLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("layout")
	tmpl, err := tmpl.Parse(templates.Layout)
	if err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Internal error: Please contact an administrator")
		return
	}
	tmpl, err = tmpl.Parse(templates.Login)
	if err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Internal error: Please contact an administrator")
		return
	}
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

	sess := db.NewSession(userToken)

	cookie := http.Cookie{
		Name:    srv.config.CookieName,
		Value:   sess.ID,
		Expires: time.Now().Add(7 * 24 * time.Hour), // TODO: Configurable expiration
		Secure:  false,
	}

	if err := srv.db.InsertSession(sess); err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "DB write failure. Please contact an administrator.")
		return
	}

	http.SetCookie(w, &cookie)
	// Redirect to form
	http.Redirect(w, r, "/", http.StatusFound)
}

func (srv *Tonic) renderForm(w http.ResponseWriter, r *http.Request, sess *db.Session) {
	tmpl := template.New("layout")
	tmpl, err := tmpl.Parse(templates.Layout)
	if err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Internal error: Please contact an administrator")
		return
	}
	tmpl, err = tmpl.Parse(templates.Form)

	if err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Internal error: Please contact an administrator")
		return
	}

	elements := make([]Element, len(srv.form))
	copy(elements, srv.form)

	data := make(map[string]interface{})
	data["elements"] = elements

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Failed to render form: %v", err)
	}
}

func (srv *Tonic) showJob(w http.ResponseWriter, r *http.Request, sess *db.Session) {
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
	if err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Internal error: Please contact an administrator")
		return
	}
	tmpl, err = tmpl.Parse(templates.Form)
	if err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Internal error: Please contact an administrator")
		return
	}

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
	timefmt := "15:04:05 Mon Jan 2 2006"
	data["submit_time"] = job.SubmitTime.Format(timefmt)
	if job.IsFinished() {
		data["end_time"] = job.EndTime.Format(timefmt)
	}
	data["messages"] = job.Messages
	if job.Error != "" {
		data["error"] = job.Error
	}
	data["readonly"] = true

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Failed to render form: %v", err)
	}
}
func (srv *Tonic) renderLog(w http.ResponseWriter, r *http.Request, sess *db.Session) {
	tmpl := template.New("layout")
	tmpl, err := tmpl.Parse(templates.Layout)
	if err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Internal error: Please contact an administrator")
		return
	}
	tmpl, err = tmpl.Parse(templates.LogView)
	if err != nil {
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Internal error: Please contact an administrator")
		return
	}

	cl := worker.NewClient(srv.config.GINServer, sess.Token)
	user, err := cl.GetSelfInfo()
	if err != nil {
		// TODO: Check for error type (unauthorized?)
		srv.web.ErrorResponse(w, http.StatusInternalServerError, "Error reading jobs from DB")
		return
	}

	joblog, err := srv.db.GetUserJobs(user.ID)
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

func (srv *Tonic) processForm(w http.ResponseWriter, r *http.Request, sess *db.Session) {
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

	client := worker.NewClient(srv.config.GINServer, sess.Token)
	srv.worker.Enqueue(worker.NewUserJob(client, jobValues))

	// redirect to job log
	http.Redirect(w, r, "/log", http.StatusSeeOther)
}
