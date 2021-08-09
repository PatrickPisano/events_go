package main

import (
	"database/sql"
	"events/pkg/storage"
	"flag"
	_ "github.com/lib/pq"
	"log"
	"os"
	"path/filepath"
)

func main() {
	dsn := flag.String("dsn", "host=localhost port=5432 user=events_user password=password dbname=events sslmode=disable", "Postgresql database connection info")
	sqlScriptPath := flag.String("sql_script_path", "./pkg/storage/postgres/.db_setup/", "sql script path")
	do := flag.String("do", "", "Action you want performed (build-tables | build-mock-data")
	flag.Parse()

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		panic(err)
	}

	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	a := app{
		db: db,
		errorLog: errorLog,
	}

	switch *do {
	case "build-tables":
		err := a.createCoreData(*sqlScriptPath)
		if err != nil {
			a.errorLog.Fatal(err)
		}
	default:
		errorLog.Fatal("invalid selection")
	}
}

type app struct {
	db *sql.DB
	errorLog *log.Logger
}

func (a app) createCoreData(sqlScriptPath string) error {
	paths := []string{
		filepath.Join(sqlScriptPath, "teardown.sql"),
		filepath.Join(sqlScriptPath, "tables.sql"),
	}
	err := storage.ExecScripts(a.db, paths...)
	if err != nil {
		return err
	}

	return nil
}