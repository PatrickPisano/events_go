package main

import (
	"database/sql"
	http2 "events/pkg/http"
	"events/pkg/services"
	"events/pkg/storage/postgres"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
)

var db *sql.DB

func main() {
	var err error

	addr := flag.String("addr", "127.0.0.1:5000", "HTTP Network Address")
	dsn := flag.String("dsn", "host=localhost port=5432 user=events_user password=password dbname=events sslmode=disable", "Postgresql database connection info")
	flag.Parse()

	db, err = sql.Open("postgres", *dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	eventRepo := postgres.NewEventStorage(db)
	eventService := services.NewEventService(eventRepo)

	app := http2.App{
		EventService: eventService,
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
