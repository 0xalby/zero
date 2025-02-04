package db

import (
	"database/sql"
	"os"

	"github.com/charmbracelet/log"
	_ "github.com/glebarez/go-sqlite"
)

// Connects to the database
func Connect(address string) (*sql.DB, error) {
	// Checks if the database exists
	if _, err := os.Stat(address); err != nil {
		// Creates one if it doesn't
		log.Infof("No database found, creating one at %s", address)
		_, err := os.Create(address)
		if err != nil {
			log.Error("failed to create database", "err", err)
			return nil, err
		}
	}
	// Opens the database using the DATABASE_DRIVER
	db, err := sql.Open(os.Getenv("DATABASE_DRIVER"), address)
	if err != nil {
		log.Error("failed to open the database", "err", err)
		return nil, err
	}
	// Pings the database to assure an established connection
	if err := db.Ping(); err != nil {
		log.Error("failed to ping the database", "err", err)
		return nil, err
	}
	return db, nil
}
