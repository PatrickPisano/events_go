package postgres

import (
	"database/sql"
	errors2 "events/pkg/errors"
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
     			link, 
    			start_time, 
    			end_time, 
    			welcome_message,
    			host_id,
    			is_published) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	row := s.DB.QueryRow(query, e.Title, e.Description, e.Link,
		e.StartTime, e.EndTime, e.WelcomeMessage, uid, false)
	var id int
	err := row.Scan(&id)
	if err != nil {
		// todo:: check for not found error
		return 0, err
	}

	return id, nil
}

func (s *EventStorage) UpdateEvent(e *events.Event) error {
	const op = "eventStorage.UpdateEvent"

	query := `UPDATE events 
					SET 
				title = $1,
				description = $2,
				link = $3, 
				start_time = $4, 
				end_time = $5, 
				welcome_message = $6, 
				is_published = $7 
			WHERE id = $8`

	_, err := s.DB.Exec(query, e.Title, e.Description, e.Link, e.StartTime, e.EndTime, e.WelcomeMessage, e.IsPublished, e.ID)
	return errors2.Wrap(err, op, "executing query")
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
	const op = "eventStorage.Event"

	// don't have to worry about sql injection because this is an integer.
	query := fmt.Sprintf("SELECT id, title FROM events WHERE id = %d", id)

	// queryrow automatically closes after it's done. You don't need to close the instance like query.
	row := s.DB.QueryRow(query)

	var e events.Event
	err := row.Scan(&e.ID, &e.Title)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("executing query [%s]: %w", op, &events.ErrNotFound{Err: err})
	} else if err != nil {
		return nil, fmt.Errorf("executing query [%s]: %w", op, err)
	}

	return &e, nil

	// err := s.DB.QueryRow(query).Scan(&e.ID, &e.Title)
}

func (s *EventStorage) SaveEventInvitations(eventID int, emails []string) error {
	const op = "eventStorage.SaveEventInvitations"

	if emails == nil {
		return nil
	}

	query := "INSERT INTO event_invitations (event_id, email) VALUES "

	var params []interface{}

	for k, email := range emails {
		query += fmt.Sprintf("(%d, $%d)", eventID, k + 1)
		if len(emails) - k > 1 {
			query += ","
		}
		params = append(params, email)
	}

	_, err := s.DB.Exec(query, params...)
	return errors2.Wrap(err, op, "executing query")
}

func (s *EventStorage) EventInvitations(id int) ([]events.Invitation, error) {
	const op = "eventStorage.EventInvitations"

	query := fmt.Sprintf("SELECT event_id, email, has_responded, accepted FROM event_invitations WHERE id = %d", id)

	rows, err := s.DB.Query(query)
	if err != nil {
		errors2.Wrap(err, op, "executing query")
	}
	defer rows.Close()

	var ii []events.Invitation
	for rows.Next() {
		var i events.Invitation
		err = rows.Scan(&i.EventID, &i.Email, &i.HasResponded, &i.Accepted)
		if err != nil {
			return nil, errors2.Wrap(err, op, "scanning into var")
		}
		ii = append(ii, i)
	}

	err = rows.Err()
	if err != nil {
		return nil, errors2.Wrap(rows.Err(), op, "checking error after scanning")
	}

	return ii, nil
}