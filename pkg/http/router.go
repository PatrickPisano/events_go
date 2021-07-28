package http

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (a *App) Routes() http.Handler {
	m := mux.NewRouter()
	m.HandleFunc("/", a.home)
	m.HandleFunc("/events", a.showEvents)
	// :[0-9]+ is regex. Makes sure it starts with a number in this case.
	m.HandleFunc("/events/{eventID:[0-9]+}", a.showEvent)
	// it only goes to this one if the request is post
	m.HandleFunc("/events/create", a.createEvent).Methods("POST")
	m.HandleFunc("/events/create", a.showEventForm)
	m.HandleFunc("/contact", a.contact)
	// gorilla mux package allows us to add the methods.
	m.HandleFunc("/register", a.register).Methods("POST")
	m.HandleFunc("/register", a.showRegistrationForm)
	m.HandleFunc("/login", a.login).Methods("POST")
	m.HandleFunc("/login", a.showLoginForm)
	m.HandleFunc("/logout", a.logout)
	m.Handle("/about", About{})

	m.Handle("/test", a.myMiddleware(http.HandlerFunc(a.test)))

	m.NotFoundHandler = http.HandlerFunc(a.notFound)


	m.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./pkg/static"))))

	return a.Session.Enable(
		a.addUserToSession(m),
		) // todo:: fix, not all request needs a session
}