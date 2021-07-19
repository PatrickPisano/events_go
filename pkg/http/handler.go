package http

import (
	"events/pkg/events"
	"events/pkg/services"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type App struct {
	EventService *services.EventService
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
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

func (a *App) showEventForm(w http.ResponseWriter, r *http.Request) {
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

		id, err := a.EventService.CreateEvent(e)
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

func (a *App) showEvents(w http.ResponseWriter, r *http.Request) {
	ee, err := a.EventService.Events()
	if err != nil {
		fmt.Println(err) // todo:: handle error
	}

	str := fmt.Sprintf("%v", ee)

	w.Write([]byte(str))
}

func (a *App) contact(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Contacts</h1>"))
}

type About struct{}

func (h About) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>About</h1>"))
}