package postgres

import (
	"database/sql"
	"events/pkg/events"
	"fmt"
)

func NewEventStorage(db *sql.DB) *EventStorage {
	return &EventStorage{db}
}

type EventStorage struct {
	DB *sql.DB
}

func (s *EventStorage) SaveEvent(e *events.Event) (int, error) {

	query := "INSERT INTO events (title) VALUES($1) RETURNING id"

	row := s.DB.QueryRow(query, e.Title)
	var id int
	err := row.Scan(&id)
	if err != nil {
		// todo:: check for not found error
		return 0, err
	}

	return id, nil
}

func (s *EventStorage) Events() ([]events.Event, error) {

	query := "SELECT id, title FROM events"

	rows, err := s.DB.Query(query)
	if err != nil {
		// todo:: check for not found error
		return nil, err
	}

	ee := make([]events.Event, 0)
	for rows.Next() {
		var e events.Event
		err = rows.Scan(&e.ID, &e.Title)
		if err != nil {
			return nil, err
		}
		ee = append(ee, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	for _, iii := range ee {
		fmt.Println(iii)
	}
	return ee, nil
}