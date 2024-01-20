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
	ErrInitDBOpen = errors.New("InitDB: open failed")
	ErrInitDBPing = errors.New("InitDB: ping failed")
)

type LoginDB struct {
	*sql.DB
}

// InitDB initializes a db connection and verifies with a Ping().
func InitDB(driverName, dataSourceName string) (*LoginDB, error) {
	// Open connection to database.
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInitDBOpen, err)
	}

	// Set desired connection parameters.
	// TODO: move values to config file
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// Ping database to confirm connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInitDBPing, err)
	}

	return &LoginDB{DB: db}, nil
}

var (
	ErrRowExistsDBNil       = errors.New("RowExists: db is nil")
	ErrRowExistsQueryFailed = errors.New("RowExists: query failed")
)

// RowExists checks if the given SQL query returns at least one row.
// The query should be in the form "SELECT 1 FROM ... WHERE ... LIMIT 1".
func (db *LoginDB) RowExists(qry string, args ...interface{}) (bool, error) {
	if db == nil {
		return false, ErrRowExistsDBNil
	}

	// QueryRow executes a query that is expected to return at most one row.
	var dummy int
	err := db.QueryRow(qry, args...).Scan(&dummy)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // Return false since No Rows.
		}
		// Unexpected error.
		return false, fmt.Errorf("%w: %v", ErrRowExistsQueryFailed, err)
	}

	return true, nil
}
