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

// InitDB initializes a db connection and verifies with a Ping().
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
// The query should be in the form "SELECT 1 FROM ... WHERE ... LIMIT 1".
func RowExists(db *sql.DB, qry string, args ...interface{}) (bool, error) {
	// QueryRow executes a query that is expected to return at most one row.
	row := db.QueryRow(qry, args...)

	// Scan and ignore the result.
	if err := row.Scan(new(interface{})); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // Return false since No Rows.
		}
		// Unexpected error.
		return false, fmt.Errorf("failed to query row: %w", err)
	}

	return true, nil
}
