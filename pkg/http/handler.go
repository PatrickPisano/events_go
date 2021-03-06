package http

import (
	"errors"
	"events/pkg/events"
	"events/pkg/services"
	"fmt"
	"github.com/golangcollege/sessions"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	sessionKeyUser string = "userID"
	sessionKeyFlash = "flash"
)

const (
	defaultMultipartFormMaxMemory = 32 << 20 // 32 * 2 ^ 20 = 32mb
)

type App struct {
	EventService *services.EventService
	UserService *services.UserService
	Session *sessions.Session
	TemplateCache map[string]*template.Template
	UploadDir string
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "home.page.tmpl", nil)
}

func (a *App) showEventForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "create-event-form.page.tmpl", nil)
}

func (a *App) createEvent(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(defaultMultipartFormMaxMemory)
	if err != nil {
		fmt.Println(err)
		// todo:: handle client error
		return
	}

	e, emails := eventFromRequest(r)

	u, ok := r.Context().Value("user").(*events.User)
	if !ok {
		a.serverError(w, r, errors.New("user was not found in the request context"))
		return
	}

	file, _, err := r.FormFile("cover_image")
	if err != nil {
		// todo:: redisplay the form with the data
		fmt.Println(err)
		return
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		// todo:: handle error
		fmt.Println(err)
		return
	}

	id, err := a.EventService.CreateEvent(e, emails, fileBytes, "jpg", u.ID)
	if err != nil {
		// todo:: handle error
		fmt.Println(err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/events/%d", id), http.StatusSeeOther)
}

func (a *App) showUpdateEventForm(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	eventID, err := strconv.Atoi(params["eventID"])
	if err != nil {
		fmt.Errorf("invalid url")
		return
	}

	e, err := a.EventService.Event(eventID)
	var notFoundErr error = &events.ErrNotFound{Err: err}
	if errors.As(err, &notFoundErr) {
		a.notFound(w, r)
		return
	} else if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.render(w, r, "update-event-form.page.tmpl", e)
}

func (a *App) updateEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["eventID"])
	if err != nil {
		a.notFound(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		fmt.Println(err)
		// todo:: handle client error
		return
	}

	e, _ := eventFromRequest(r)
	e.ID = id

	err = a.EventService.UpdateEvent(e)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.Session.Put(r, sessionKeyFlash, "Event successfully updated!") // todo:: save flash value

	http.Redirect(w, r, fmt.Sprintf("/events/%d", id), http.StatusSeeOther)
}

func (a *App) showEvents(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value("user").(*events.User)
	if !ok  {
		a.serverError(w, r, errors.New("user not found in request context"))
		return
	}

	ee, err := a.EventService.Events(u.ID)
	if err != nil {
		fmt.Println(err) // todo:: handle error
	}

	a.render(w, r, "event-list.page.tmpl", ee)
}

func (a *App) showEvent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["eventID"])
	if err != nil {
		fmt.Println(err) // todo:: handle error
		return
	}

	e, err := a.EventService.Event(id)
	if err != nil {
		notFound := &events.ErrNotFound{}
		if errors.As(err, &notFound) {
			a.notFound(w, r)
			return
		}

		a.serverError(w, r, err)
		return
	}

	a.render(w, r, "event-detail.page.tmpl", struct {
		Event *events.Event
	}{e})
}

func (a *App) contact(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Contacts</h1>"))
}

func (a *App) showRegistrationForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "register.page.tmpl", nil)
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// todo:: display client error
		fmt.Println(err)
		return
	}

	password := r.PostForm.Get("password")

	u := &events.User{
		Names:   r.PostForm.Get("names"),
		Email:   r.PostForm.Get("email"),
	}

	_, err = a.UserService.CreateUser(u, password)
	if err != nil {
		conflictErr := &events.ErrConflict{}
		if errors.As(err, &conflictErr) {
			a.render(w, r, "register.page.tmpl", struct {
				DuplicateEmail bool
				Names string
				Email       string
			}{true, u.Names,u.Email})
			return
		}
		a.serverError(w, r, err)
		return
	}

	// 303 good for when people need to be redirected after signing up, etc
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *App) showLoginForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "login.page.tmpl", nil)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// todo:: display client error
		fmt.Println(err)
		return
	}

	returnURL := r.FormValue("return_url")
	if returnURL == "" {
		returnURL = "/"
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	match, id, err := a.UserService.EmailMatchPassword(email, password)
	if err != nil {
		notFoundErr := &events.ErrNotFound{}
		if errors.As(err, &notFoundErr) {
			a.render(w, r, "login.page.tmpl", struct{
				LoginFailed bool
				Email string
			}{true, email})
			return
		}
		a.serverError(w ,r, err)
		return
	}

	// email valid but didn't match password
	if !match {
		a.render(w, r, "login.page.tmpl", struct{
			LoginFailed bool
			Email string
		}{true, email})
		return
	}

	a.Session.Put(r, sessionKeyUser, id)

	http.Redirect(w, r, returnURL, http.StatusSeeOther)
	// 303 good for when people need to be redirected after signing up, etc
	// http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	a.Session.Remove(r, sessionKeyUser)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type About struct{}

func (h About) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>About</h1>"))
}

func (a *App) test(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Test handler called")
	w.Write([]byte("Test handler called"))
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Page was not found"))
}

func (a *App) serverError(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func NewTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Use the filepath.Glob function to get a slice of all filepaths with
	// the extension '.page.tmpl'. This essentially gives us a slice of all the
	// 'page' templates for the application.
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		// Extract the file name (like 'home.page.tmpl') from the full file path
		// and assign it to the name variable.
		name := filepath.Base(page)

		ts, err := template.New(name).ParseFiles(page) // todo:: add .Func(...).
		if err != nil {
			return nil, err
		}

		// Use the ParseGlob method to add any 'layout' templates to the
		// template set.
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		// Use the ParseGlob method to add any 'partial' templates to the
		// template set.
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		// Add the template set to the cache, using the name of the page
		// (like 'home.page.tmpl') as the key.
		cache[name] = ts
	}

	return cache, nil
}

func (a App) render(w http.ResponseWriter, r *http.Request, name string, td interface{}) {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.tmpl'). If no entry exists in the cache with the
	// provided name, call the serverError helper.
	ts, ok := a.TemplateCache[name]
	if !ok {
		err := fmt.Errorf("server error: the template %s does not exist", name)
		a.serverError(w, r, err)
	}

	// Execute the template set, passing in any dynamic data.
	err := ts.Execute(w, a.addDefaultData(r, td)) // todo:: add default data
	if err != nil {
		a.serverError(w, r, err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

type templateData struct {
	User *events.User
	Flash string
	CurrentYear int
	Data interface{}
}

func (a *App) addDefaultData(r *http.Request, data interface{}) *templateData {
	//u, _ := events.UserFromContext(r.Context())

	u, _ := r.Context().Value("user").(*events.User)
	fmt.Println(u)

	td := &templateData{
		User: u,
		Flash: a.Session.PopString(r, sessionKeyFlash),
		//CurrentYear: time.Now().Year(),
		Data: data,
	}

	return td
}

func eventFromRequest(r *http.Request) (*events.Event, []string) {
	var err error
	var layout = "2006-01-02T15:04"

	var startTime time.Time
	startTimeStr := r.PostForm.Get("start_time")
	if startTimeStr != "" {
		startTime, err = time.Parse(layout, startTimeStr)
		if err != nil {
			// todo:: handle error
			panic(err)
		}
	}

	var stopTime time.Time
	stopTimeStr := r.PostForm.Get("stop_time")
	if stopTimeStr != "" {
		startTime, err = time.Parse(layout, stopTimeStr)
		if err != nil {
			// todo:: handle error
			panic(err)
		}
	}

	e := &events.Event{
		Title: r.PostForm.Get("title"),
		Description: r.PostForm.Get("description"),
		Link: r.PostForm.Get("link"),
		StartTime: &startTime,
		EndTime: &stopTime,
		WelcomeMessage: r.PostForm.Get("welcome_message"),
		IsPublished: false,
	}

	emailsStr := r.PostForm.Get("invitations")         // format " email1@example.com, email2@example.com,email3@example.com"
	emailsStr = strings.ReplaceAll(emailsStr, " ", "") // new format "email1@example.com,email2@example.com,email3@example.com"
	emails := strings.Split(emailsStr, ",") // new data ["email1@example.com", "email2@example.com", "email3@example.com"]

	return e, emails
}