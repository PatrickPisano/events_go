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

func (s *EventStorage) SaveEvent(e *events.Event, uid int) (int, error) {

	query := `INSERT INTO events 
    			(title, 
     			description, 
     			is_virtual, 
     			address, 
     			link, 
     			number_of_seats, 
    			start_time, 
    			end_time, 
    			welcome_message,
    			host_id,
    			is_published) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	row := s.DB.QueryRow(query, e.Title, e.Description, e.IsVirtual, e.Address, e.Link, e.NumberOfSeats,
		e.StartTime, e.EndTime, e.WelcomeMessage, uid, false)
	var id int
	err := row.Scan(&id)
	if err != nil {
		// todo:: check for not found error
		return 0, err
	}

	return id, nil
}

func (s *EventStorage) Events(uid int) ([]events.Event, error) {

	query := fmt.Sprintf("SELECT id, title FROM events WHERE host_ID = %d", uid)

	rows, err := s.DB.Query(query)
	if err != nil {
		// todo:: check for not found error
		return nil, err
	}
	// any time you use query, you have to close the instance. Important!
	defer rows.Close()

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

func (s *EventStorage) Event(id int) (*events.Event, error) {
	// don't have to worry about sql injection because this is an integer.
	query := fmt.Sprintf("SELECT id, title FROM events WHERE id = %d", id)

	// queryrow automatically closes after it's done. You don't need to close the instance like query.
	row := s.DB.QueryRow(query)

	var e events.Event
	err := row.Scan(&e.ID, &e.Title)
	if err != nil {
		return nil, err
	}

	return &e, nil

	// err := s.DB.QueryRow(query).Scan(&e.ID, &e.Title)
}