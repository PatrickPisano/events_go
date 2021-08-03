package http

import (
	"errors"
	"events/pkg/events"
	"events/pkg/services"
	"fmt"
	"github.com/golangcollege/sessions"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type App struct {
	EventService *services.EventService
	UserService *services.UserService
	Session *sessions.Session
	TemplateCache map[string]*template.Template
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "home.page.tmpl", nil)
}

func (a *App) showEventForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "create-event-form.page.tmpl", nil)
}

func (a *App) createEvent(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		// todo:: handle client error
		return
	}

	var isVirtual bool

	if t := r.PostForm.Get("is_virtual"); t == "virtual" {
		isVirtual = true
	} else if t == "physical" {
		isVirtual = false
	} else {
		// todo:: return an error
		panic("unexpected type for the 'is_virtual' field")
	}

	var numOfSeats int

	numOfSeatsStr := r.PostForm.Get("number_of_seats")
	if numOfSeatsStr != "" {
		numOfSeats, err = strconv.Atoi(numOfSeatsStr)
		if err != nil {
			// todo:: return an error
			panic(err)
		}
	}

	layout := "2006-01-02T15:04"

	var startTime time.Time

	startTimeStr := r.PostForm.Get("start_time")
	if startTimeStr != "" {
		startTime, err = time.Parse(layout, startTimeStr)
		if err != nil {
			// todo:: return an error
			panic(err)
		}
	}

	var endTime time.Time

	endTimeStr := r.PostForm.Get("end_time")
	if endTimeStr != "" {
		endTime, err = time.Parse(layout, endTimeStr)
		if err != nil {
			// todo:: return an error
			panic(err)
		}
	}

	u, ok := r.Context().Value("user").(*events.User)
	if !ok {
		a.serverError(w, r, errors.New("user was not found in the request context"))
		return
	}

	e := &events.Event{
		// use PostForm to only get values from post (not get - the url)
		Title:          r.PostForm.Get("title"),
		Description:    r.PostForm.Get("description"),
		IsVirtual:      isVirtual,
		Address:        r.PostForm.Get("address"),
		Link:           r.PostForm.Get("link"),
		NumberOfSeats:  numOfSeats,
		StartTime:      startTime,
		EndTime:        endTime,
		WelcomeMessage: r.PostForm.Get("welcome_message"),
		HostID:	 		u.ID,
		IsPublished:    false,
	}

	id, err := a.EventService.CreateEvent(e, u.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

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
		fmt.Println(err) // todo:: handle error
	}

	a.render(w, r, "event-detail.page.tmpl", e)
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
		//todo:: check error due to duplicate emails
		fmt.Println("error", err)
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

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	match, id, err := a.UserService.EmailMatchPassword(email, password)
	if err != nil {
		//todo:: check error due to duplicate emails
		fmt.Println("error", err)
		return
	}

	if match {
		a.Session.Put(r, "userID", id) // todo:: use a const to save keys
		w.Write([]byte("Logged in"))
	} else {
		w.Write([]byte("Login was not successful"))
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	// 303 good for when people need to be redirected after signing up, etc
	// http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	a.Session.Remove(r, "userID")
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
		/*ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}*/

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
		//Flash: a.Session.PopString(r, "flash")
		//CurrentYear: time.Now().Year()
		Data: data,
	}

	return td
}