// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrDBOpen = errors.New("failed to open db")
	ErrDBPing = errors.New("failed to ping db")
)

// InitDB initializes a connection to the database with given driver and source names.
// It sets the database connection parameters and verifies the connection via ping.
func InitDB(driverName, dataSourceName string) (*sql.DB, error) {
	// open connection to database
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBOpen, err)
	}

	// set desire connection parameters
	// TODO: move values to config file
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// ping database to confirm connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBPing, err)
	}

	return db, err
}

// RowExists checks if the given SQL query returns at least one row.
// The query should be in the form "SELECT 1 FROM ... WHERE ...".
func RowExists(db *sql.DB, qry string, args ...interface{}) (bool, error) {
	var num int

	row := db.QueryRow(qry, args...)
	if err := row.Scan(&num); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("error checking row existence: %w", err)
	}

	return true, nil
}
