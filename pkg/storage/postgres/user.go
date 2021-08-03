package postgres

import (
	"database/sql"
	"events/pkg/events"
	"fmt"
)

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db}
}

type UserStorage struct {
	DB *sql.DB
}

func (s *UserStorage) SaveUser(e *events.User, password string) (int, error) {
	query := `INSERT INTO users 
    			(names, 
     			 email,
    			 password
     			) VALUES($1, $2, $3) RETURNING id`

	row := s.DB.QueryRow(query, e.Names, e.Email, password)
	var id int
	err := row.Scan(&id)
	if err != nil {
		// todo:: check for not found error
		return 0, err
	}

	return id, nil
}

func (s *UserStorage) UserIDAndPasswordByEmail(email string) (int, string, error) {
	query := "SELECT id, password FROM users WHERE email=$1"

	row := s.DB.QueryRow(query, email)

	var id int
	var password string

	err := row.Scan(&id, &password)
	if err != nil {
		// todo:: check for not found error
		return 0, "", err
	}

	return id, password, nil
}

func (s *UserStorage) User(uid int) (*events.User, error) {
	query := fmt.Sprintf("SELECT names, email FROM users WHERE id=%v", uid)

	row := s.DB.QueryRow(query)

	var u events.User
	err := row.Scan(&u.Names, &u.Email)

	// golang convention to use Err to start the Err name
	if err == sql.ErrNoRows {
		// todo:: return not found
		return nil, err
	} else if err != nil {
		return nil, err
	}

	u.ID = uid

	return &u, nil
}

// SELECT * FROM users WHERE email='' AND password=''
// SELECT id, password FROM users WHERE email='' AND password=''

// SELECT * FROM users WHERE id=''
// SELECT names, email FROM users WHERE id=''