package postgres

import (
	"database/sql"
	errors2 "events/pkg/errors"
	"events/pkg/events"
	"fmt"
	"github.com/lib/pq"
)

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db}
}

type UserStorage struct {
	DB *sql.DB
}

func (s *UserStorage) SaveUser(e *events.User, password string) (int, error) {
	const op = "userStorage.SaveUser"

	query := `INSERT INTO users 
    			(names, 
     			 email,
    			 password
     			) VALUES($1, $2, $3) RETURNING id`

	row := s.DB.QueryRow(query, e.Names, e.Email, password)
	var id int
	err := row.Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case /*"foreign_key_violation",*/ "unique_violation":
				return 0, errors2.Wrap(&events.ErrConflict{Err: pqErr}, op, "executing query")
			}
		}
		return 0, errors2.Wrap(err, op, "executing query")
	}

	return id, nil
}

func (s *UserStorage) UserIDAndPasswordByEmail(email string) (int, string, error) {
	const op = "userStorage.UserIDAndPasswordByEmail"

	query := "SELECT id, password FROM users WHERE email=$1"

	row := s.DB.QueryRow(query, email)

	var id int
	var password string

	err := row.Scan(&id, &password)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", errors2.Wrap(&events.ErrNotFound{Err: err}, op, "executing query")
		}
		return 0, "", errors2.Wrap(&events.ErrNotFound{Err: err}, op, "executing query")
	}

	return id, password, nil
}

func (s *UserStorage) User(uid int) (*events.User, error) {
	const op = "userStorage.User"

	query := fmt.Sprintf("SELECT names, email FROM users WHERE id=%v", uid)

	row := s.DB.QueryRow(query)

	var u events.User
	err := row.Scan(&u.Names, &u.Email)

	// golang convention to use Err to start the Err name
	if err == sql.ErrNoRows {
		return nil, errors2.Wrap(&events.ErrNotFound{Err: err}, op, "executing query")
	} else if err != nil {
		return nil, errors2.Wrap(err, op, "executing query")
	}

	u.ID = uid

	return &u, nil
}

// SELECT * FROM users WHERE email='' AND password=''
// SELECT id, password FROM users WHERE email='' AND password=''

// SELECT * FROM users WHERE id=''
// SELECT names, email FROM users WHERE id=''