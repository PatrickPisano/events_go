package postgres

import (
	"database/sql"
	"errors"
	errors2 "events/pkg/errors"
	"events/pkg/events"
	"events/pkg/storage"
	"fmt"
	"os"
	"path/filepath"
)

func NewEventStorage(db *sql.DB, uploadDir string) *EventStorage {
	return &EventStorage{db, uploadDir}
}

type EventStorage struct {
	DB *sql.DB
	UploadDir string
}

func (s *EventStorage) Tx() (*sql.Tx, error) {
	return s.DB.Begin()
}

func (s *EventStorage) SaveEventTx(tx *sql.Tx, e *events.Event, uid int) (int, error) {
	query := `INSERT INTO events 
    			(title, 
     			description,  
     			link, 
    			start_time, 
    			end_time, 
    			welcome_message,
    			cover_image_path,
    			host_id,
    			is_published) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	row := tx.QueryRow(query, e.Title, e.Description, e.Link,
		e.StartTime, e.EndTime, e.WelcomeMessage, e.CoverImagePath, uid, false)
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
				cover_image_path = $7,
				is_published = $8 
			WHERE id = $9`

	_, err := s.DB.Exec(query, e.Title, e.Description, e.Link, e.StartTime, e.EndTime, e.WelcomeMessage, e.CoverImagePath, e.IsPublished, e.ID)
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

	return ee, nil
}

func (s *EventStorage) UpdateEventTx(tx *sql.Tx, i *events.Event) error {
	const op = "eventStorage.UpdateEventTx"

	if tx == nil {
		return errors2.Wrap(errors.New("tx is nil"), op, "checking transaction")
	}

	query := `UPDATE events 
				SET 
				    title = $1, 
				    description = $2, 
				    link = $3, 
				    start_time = $4, 
				    end_time = $5, 
				    welcome_message = $6, 
				    is_published = $7, 
				    cover_image_path = $8 
				WHERE id = $9`

	_, err := tx.Exec(query, i.Title, i.Description, i.Link,
		i.StartTime, i.EndTime, i.WelcomeMessage, i.IsPublished, i.CoverImagePath, i.ID)
	if err != nil {
		return errors2.Wrap(err, op, "executing query")
	}

	return nil
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

func (s *EventStorage) SaveEventCoverImage(image []byte, key []string, ext string) error {
	const op = "eventStorage.SaveEventCover"

	fullPath := []string{s.UploadDir}
	var uniquePath []string
	for _, v := range key[:len(key) - 1] {
		uniquePath = append(uniquePath, v)
	}
	fullPath = append(fullPath, uniquePath...)

	// create random directory if not exist
	if _, err := os.Stat(filepath.Join(fullPath...)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Join(fullPath...), os.ModeDir)
		if err != nil {
			return err
		}
	}

	// return error if file exists
	_, err := os.Stat(filepath.Join(filepath.Join(fullPath...), key[2:len(key)][0] + "." + ext))
	if err == nil { // file exists, return error
		return errors2.Wrap(errors.New("file with name already exist"), op, "checking if file exists")
	} else if !os.IsNotExist(err) { // error is not NotExist, return error
		if err != nil {
			return err
		}
	}

	// create empty file
	file2, err := os.OpenFile(
		filepath.Join(filepath.Join(fullPath...),
			key[2:len(key)][0] + "." + ext),
		os.O_WRONLY|os.O_CREATE,
		os.ModeDir,
	)
	if err != nil {
		return err
	}
	defer file2.Close()

	// copy uploaded file byte into newly  created empty file
	_, err = file2.Write(image)
	if err != nil {
		return err
	}

	return nil
}


/*func (s *EventStorage) SaveEventCoverImage(image []byte, key []string, ext string) error {
	const op = "eventStorage.SaveEventCoverImage"

	f, err := os.Create(filepath.Join(s.UploadDir, fmt.Sprintf("%d.jpg", time.Now().UnixNano())))
	if err != nil {
		return errors2.Wrap(err, op, "creating empty image file")
	}

	_, err = f.Write(image)
	if err != nil {
		// todo:: handle error
		fmt.Println(err)
		return errors2.Wrap(err, op, "writing read bytes to file")
	}

	return nil
}*/

func (s *EventStorage) event(db storage.DB, id int) (*events.Event, error) {
	const op = "eventStorage.event"

	if db == nil {
		return nil, errors2.Wrap(errors.New("db is nil"), op, "fetching event")
	}

	query := fmt.Sprintf(
		`SELECT id, title, description, link, start_time, end_time, welcome_message, cover_image_path, is_published, host_id FROM events WHERE id = %d`,
		id,
		)

	row := db.QueryRow(query)

	var description, link, welcomeMessage, coverImagePath sql.NullString
	var startTime, endTime sql.NullTime

	var e events.Event
	err := row.Scan(&e.ID, &e.Title, &description, &link, &startTime,
		&endTime, &welcomeMessage, &coverImagePath, &e.IsPublished, &e.HostID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors2.Wrap(&events.ErrNotFound{Err: err}, op, "scanning into var")
		}
		return nil, errors2.Wrap(err, op, "scanning into var")
	}

	e.Description = storage.NullableStrToStr(description)
	e.Link = storage.NullableStrToStr(link)
	e.WelcomeMessage = storage.NullableStrToStr(welcomeMessage)
	e.CoverImagePath = storage.NullableStrToStr(coverImagePath)
	e.StartTime = storage.SqlTimeToTime(startTime)
	e.EndTime = storage.SqlTimeToTime(endTime)

	return &e, nil
}

func (s *EventStorage) Event(id int) (*events.Event, error) {
	const op = "eventStorage.Event"

	e, err := s.event(s.DB, id)
	return e, errors2.Wrap(err, op, "getting event")
}

func (s *EventStorage) EventTx(tx *sql.Tx, id int) (*events.Event, error) {
	const op = "eventStorage.EventTx"

	e, err := s.event(tx, id)
	return e, errors2.Wrap(err, op, "getting event")
}

func (s *EventStorage) SaveEventInvitationsTx(tx *sql.Tx, eventID int, invitations []events.Invitation) error {
	const op = "eventStorage.SaveEventInvitationsTX"

	if tx == nil {
		return errors.New("tx is nil")
	}

	if invitations == nil {
		return nil
	}

	query := "INSERT INTO event_invitations (event_id, email, token) VALUES "

	var params []interface{}

	counter := 0
	for k, invitation := range invitations {
		counter++
		emailPlaceholder := counter
		counter++
		tokenPlaceholder := counter
		query += fmt.Sprintf("(%d, $%d, $%d)", eventID, emailPlaceholder, tokenPlaceholder) // (2, $1, $2), (2, $3, $4)
		if len(invitations) - k > 1 {
			query += ","
		}
		params = append(params, invitation.Email, invitation.Token)
	}

	_, err := tx.Exec(query, params...)
	return errors2.Wrap(err, op, "executing query")
}

func (s *EventStorage) DeleteAllEventInvitationsTx(tx *sql.Tx, eventID int) error {
	const op = "eventStorage.DeleteAllEventInvitationsTx"

	if tx == nil {
		return errors.New("tx is nil")
	}

	query := fmt.Sprintf("DELETE FROM event_invitations WHERE event_id = %d", eventID)

	_, err := tx.Exec(query)
	return errors2.Wrap(err, op, "executing query")
}
