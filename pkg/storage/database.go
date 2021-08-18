package storage

import (
	"database/sql"
	"io/ioutil"
	"strings"
	"time"
)

type DB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// ExecScripts will receive slice of paths (string) and execute all sql
// statements in it in the order in which they are passed
func ExecScripts(db *sql.DB, paths ...string) error {
	for _, p := range paths {
		bb, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}

		if strings.TrimSpace(string(bb)) == "" {
			break
		}

		_, err = db.Exec(string(bb))
		if err != nil {
			return err
		}
	}

	return nil
}

func NullableStrToStr(s1 sql.NullString) string {
	var s2 string
	if !s1.Valid {
		return s2
	}

	return s1.String
}

// SqlTimeToTime takes sq.NullTime and returns corresponding *time.Time
// If time is not zero time, it returns it otherwise it returns a zero time
func SqlTimeToTime(t1 sql.NullTime) *time.Time {
	var t2 *time.Time
	if !t1.Valid {
		return t2
	}

	return &t1.Time
}