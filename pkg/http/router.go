package http

import "net/http"

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", a.home)
	mux.HandleFunc("/events", a.showEvents)
	mux.HandleFunc("/events/create", a.showEventForm)
	mux.HandleFunc("/contact", a.contact)
	mux.Handle("/about", About{})

	mux.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("./pkg/static"))),
	)

	return mux
}