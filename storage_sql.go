package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SQLStorage is a SQLite backed storage adapter.
type SQLStorage struct {
	db        *sql.DB
	filesRoot string
}

// NewSQLStorage returns a new SQL Storage.
func NewSQLStorage(dbFilename, filesRoot string) Storage {
	db, err := sql.Open("sqlite", dbFilename)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS apartments
		(
			id INTEGER PRIMARY KEY
			,address TEXT
			,floor REAL
			,area INTEGER
			,rooms REAL
			,price INTEGER
			,estimatedValue INTEGER
			,fee INTEGER
		)
	`)
	if err != nil {
		panic(err)
	}

	return &SQLStorage{
		db:        db,
		filesRoot: filesRoot,
	}
}

// Put stores an apartment.
func (s *SQLStorage) Put(apt Apartment) error {
	_, err := s.db.Exec(`
		REPLACE INTO apartments
		(
			id
			,address
			,floor
			,area
			,rooms
			,price
			,estimatedValue
			,fee
		) VALUES (
			?
			,?
			,?
			,?
			,?
			,?
			,?
			,?
		)
	`, apt.ID,
		apt.Address,
		apt.Floor,
		apt.Area,
		apt.Rooms,
		apt.Price,
		apt.EstimatedValue,
		apt.Fee,
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if err := ensureDir(s.filesRoot, 0775); err != nil {
		return fmt.Errorf("ensure root dir: %w", err)
	}

	for i, url := range apt.ImageURLs {
		downloaded, err := downloadImage(url, s.filesRoot)
		if err != nil {
			return fmt.Errorf("download image: %w", err)
		}

		if !downloaded {
			fmt.Printf("Skipped %s (%d/%d)\n", url, i+1, len(apt.ImageURLs))
			continue
		}
		fmt.Printf("Downloaded %s (%d/%d)\n", url, i+1, len(apt.ImageURLs))
	}
	return nil
}
