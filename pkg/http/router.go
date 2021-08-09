package http

import (
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"net/http"
)

func (a *App) Routes() http.Handler {
	sessionMiddleware := alice.New(a.Session.Enable, a.addUserToSession)
	authenticatedOnly := alice.New(sessionMiddleware.Then, a.authenticatedUser)

	m := mux.NewRouter()
	m.Handle("/", sessionMiddleware.Then(http.HandlerFunc(a.home))).Methods("GET")
	m.Handle("/events", authenticatedOnly.Then(http.HandlerFunc(a.showEvents))).Methods("GET")
	// :[0-9]+ is regex. Makes sure it starts with a number in this case.
	m.Handle("/events/{eventID:[0-9]+}", sessionMiddleware.Then(http.HandlerFunc(a.showEvent))).Methods("GET")
	// it only goes to this one if the request is post
	m.Handle("/events/create", authenticatedOnly.Then(http.HandlerFunc(a.createEvent))).Methods("POST")
	m.Handle("/events/create", authenticatedOnly.Then(http.HandlerFunc(a.showEventForm))).Methods("GET")

	m.Handle("/events/{eventID:[0-9]+}/edit", authenticatedOnly.Then(http.HandlerFunc(a.updateEvent))).Methods("POST")
	m.Handle("/events/{eventID:[0-9]+}/edit", authenticatedOnly.Then(http.HandlerFunc(a.showUpdateEventForm))).Methods("GET")
	m.Handle("/contact", sessionMiddleware.Then(http.HandlerFunc(a.contact))).Methods("GET")
	// gorilla mux package allows us to add the methods.
	m.Handle("/register", sessionMiddleware.Then(http.HandlerFunc(a.register))).Methods("POST")
	m.Handle("/register", sessionMiddleware.Then(http.HandlerFunc(a.showRegistrationForm))).Methods("GET")
	m.Handle("/login", sessionMiddleware.Then(http.HandlerFunc(a.login))).Methods("POST")
	m.Handle("/login", sessionMiddleware.Then(http.HandlerFunc(a.showLoginForm))).Methods("GET")
	m.Handle("/logout", sessionMiddleware.Then(http.HandlerFunc(a.logout))).Methods("GET")
	m.Handle("/about", About{})

	m.Handle("/test", a.myMiddleware(http.HandlerFunc(a.test)))

	m.NotFoundHandler = http.HandlerFunc(a.notFound)


	m.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./pkg/static"))))

	return m
}