package main

import (
	"database/sql"
	"events/pkg/events"
	"events/pkg/services"
	"events/pkg/storage/postgres"
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"net/http"
	"strings"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=events_user password=password dbname=events sslmode=disable")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", home)
	mux.HandleFunc("/events", showEvents)
	mux.HandleFunc("/events/create", showEventForm)
	mux.HandleFunc("/contact", contact)
	mux.Handle("/about", About{})

	mux.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("./pkg/static"))),
			)

	serv := &http.Server{
		Addr:    ":5000",
		Handler: mux,
	}

	fmt.Println("Server started at :5000")

	err = serv.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

func home(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("./pkg/static/templates/home.page.tmpl")
	if err != nil {
		fmt.Println("error", err)
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		fmt.Println("error", err)
		return
	}
}

func showEventForm(w http.ResponseWriter, r *http.Request) {
	if strings.ToLower(r.Method) == "post" {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			// todo:: handle client error
			return
		}

		e := &events.Event{
			// use PostForm to only get values from post (not get - the url)
			Title:          r.PostForm.Get("title"),
			Description:    r.PostForm.Get("description"),
			IsVirtual:      false,
			Address:        r.PostForm.Get("address"),
			Link:           r.PostForm.Get("link"),
			NumberOfSeats:  0,
			StartTime:      nil,
			EndTime:        nil,
			WelcomeMessage: r.PostForm.Get("welcome_message"),
			IsPublished:    false,
		}

		eventRepo := postgres.NewEventStorage(db)
		eventService := services.NewEventService(eventRepo)

		id, err := eventService.CreateEvent(e)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("event id:", id)
	} else {
		ts, err := template.ParseFiles("./pkg/static/templates/create-event-form.page.tmpl")
		if err != nil {
			fmt.Println("error", err)
			return
		}

		err = ts.Execute(w, nil)
		if err != nil {
			fmt.Println("error", err)
			return
		}
	}
}

func showEvents(w http.ResponseWriter, r *http.Request) {
	eventRepo := postgres.NewEventStorage(db)
	eventService := services.NewEventService(eventRepo)

	ee, err := eventService.Events()
	if err != nil {
		fmt.Println(err) // todo:: handle error
	}

	str := fmt.Sprintf("%v", ee)

	w.Write([]byte(str))
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Contacts</h1>"))
}

type About struct{}

func (h About) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>About</h1>"))
}
