package main

import (
	"database/sql"
	http2 "events/pkg/http"
	"events/pkg/services"
	"events/pkg/storage/postgres"
	"flag"
	"fmt"
	"github.com/golangcollege/sessions"
	_ "github.com/lib/pq"
	"net/http"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:5000", "HTTP Network Address")
	dsn := flag.String("dsn", "host=localhost port=5432 user=events_user password=password dbname=events sslmode=disable", "Postgresql database connection info")
	flag.Parse()

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	eventRepo := postgres.NewEventStorage(db)
	eventService := services.NewEventService(eventRepo)

	userRepo := postgres.NewUserStorage(db)
	userService := services.NewUserService(userRepo)

	session := sessions.New([]byte("secret"))
	// might want to set this to a longer time in production
	session.Lifetime = time.Hour * 12

	templateCache, err := http2.NewTemplateCache("./pkg/static/templates")
	if err != nil {
		panic(err)
	}

	app := http2.App{
		EventService: eventService,
		UserService: userService,
		Session: session,
		TemplateCache: templateCache,
	}

	serv := &http.Server{
		Addr:    *addr,
		Handler: app.Routes(),
	}

	fmt.Println("Server started at ", *addr)

	err = serv.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
