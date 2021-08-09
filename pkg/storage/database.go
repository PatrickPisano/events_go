package storage

import (
	"database/sql"
	"io/ioutil"
	"strings"
)

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