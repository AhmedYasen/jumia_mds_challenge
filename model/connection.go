package model

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var dbConn *sql.DB

func ConnectDatabase() error {
	db, err := sql.Open("sqlite3", "./mds.db")
	if err != nil {
		return err
	}

	dbConn = db
	return nil
}
