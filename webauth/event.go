// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

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
	EventLogout    EventName = "logout"
	EventRegister  EventName = "register"
	EventSaveToken EventName = "save_token"
	EventResetPass EventName = "reset_pass"
	EventConfirmed EventName = "confirmed"
	EventMaxName   EventName = "1234567890" // Event defined as varchar(10).
)

// Event represents a system event, such as a user login or registration.
type Event struct {
	Name      EventName // Name of the event.
	Succeeded bool      // Indicates if the event was successful.
	Username  string    // Username associated with the event.
	Message   string    // Message or details about the event.
	Created   time.Time // Timestamp of event, set by db when inserted.
}

// LogValue implements slog.LogValuer. It returns a group containing
// the fields of Event, so that they appear together in the log output.
func (e Event) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("Name", string(e.Name)),
		slog.Bool("Succeeded", e.Succeeded),
		slog.String("Username", e.Username),
		slog.String("Message", e.Message),
		slog.Time("Created", e.Created),
	)
}

var (
	ErrWriteEventDBNil  = errors.New("WriteEvent: db is nil")
	ErrWriteEventFailed = errors.New("WriteEvent: db write failed")
)

// WriteEvent saves an event to database.
func (db *AuthDB) WriteEvent(name EventName, succeeded bool, username, message string) error {
	e := Event{Name: name, Succeeded: succeeded, Username: username, Message: message}
	logger := slog.With("event", e, "func", "WriteEvent")

	if db == nil {
		logger.Error("nil db")
		return ErrWriteEventDBNil
	}

	const qry = `INSERT INTO events(name, succeeded, username, message) VALUES(?, ?, ?, ?)`
	result, err := db.Exec(qry, e.Name, e.Succeeded, e.Username, e.Message)
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

var (
	ErrGetEventsDBNil = errors.New("GetEvents: db is nil")
	ErrGetEventsQuery = errors.New("GetEvents: query failed")
	ErrGetEventsScan  = errors.New("GetEvents: scan failed")
	ErrGetEventsRows  = errors.New("GetEvents: rows.Err()")
)

// GetEvents returns a list of all events.
func (db *AuthDB) GetEvents() ([]Event, error) {
	logger := slog.With("func", "GetEvents")

	if db == nil {
		logger.Error("db is nil")
		return nil, ErrRowExistsDBNil
	}

	qry := `SELECT name, succeeded, username, message, created FROM events ORDER BY created DESC`
	rows, err := db.Query(qry)
	if err != nil {
		slog.Error("query for events failed", "err", err)
		return nil, fmt.Errorf("%w: %v", ErrGetEventsQuery, err)

	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event

		err := rows.Scan(&event.Name, &event.Succeeded, &event.Username, &event.Message, &event.Created)
		if err != nil {
			slog.Error("failed rows.Scan", "err", err)
			return nil, fmt.Errorf("%w: %v", ErrGetEventsScan, err)
		}

		events = append(events, event)
	}

	err = rows.Err()
	if err != nil {
		slog.Error("failed rows.Err", "err", err)
		return nil, fmt.Errorf("%w: %v", ErrGetEventsRows, err)
	}

	return events, nil
}
