// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// EventName represents possible event types within the system.
type EventName string

const (
	EventLogin     EventName = "login"
	EventLogout              = "logout"
	EventRegister            = "register"
	EventSaveToken           = "save_token"
	EventResetPass           = "reset_pass"
	EventConfirmed           = "confirmed"
	EventMaxName             = "1234567890" // Event defined as varchar(10).
)

// Event represents a system event, such as a user login or registration.
type Event struct {
	Name     EventName // Name of the event.
	Success  bool      // Indicates if the event was successful or not.
	Username string    // Username associated with the event.
	Message  string    // Message or details about the event.
	Created  time.Time // Timestamp of event, set by db when inserted.
}

// LogValue implements slog.LogValuer to return a group of the Event fields.
func (e Event) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("Name", string(e.Name)),
		slog.Bool("Success", e.Success),
		slog.String("Username", e.Username),
		slog.String("Message", e.Message),
		slog.Time("Created", e.Created),
	)
}

var (
	ErrWriteEventNilDB  = errors.New("nil db handle")
	ErrWriteEventFailed = errors.New("failed to write event to db")
)

// WriteEvent saves an event to database.
func (db *LoginDB) WriteEvent(name EventName, success bool, username, message string) error {
	e := Event{Name: name, Success: success, Username: username, Message: message}
	logger := slog.With("event", e, "func", "WriteEvent")

	if db == nil {
		logger.Error("nil db")
		return ErrWriteEventNilDB
	}

	const qry = `INSERT INTO events(name, success, username, message) VALUES(?, ?, ?, ?)`
	result, err := db.Exec(qry, e.Name, e.Success, e.Username, e.Message)
	if err != nil {
		logger.Error("failed to write event", "err", err)
		return fmt.Errorf("%w: %v", ErrWriteEventFailed, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("failed to get rows affected", "err", err)
		return fmt.Errorf("%w: %v", ErrWriteEventFailed, err)
	}

	if rowsAffected != 1 {
		logger.Error("number of rows affected is not one",
			"rows", rowsAffected)
		return fmt.Errorf("%w: %v", ErrWriteEventFailed, err)
	}

	logger.Debug("wrote event to database")
	return nil
}
