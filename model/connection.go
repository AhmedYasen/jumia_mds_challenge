package model

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var dbConn *sql.DB

func ConnectDatabase(path string) error {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	dbConn = db
	return nil
}
