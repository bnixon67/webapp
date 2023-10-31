package weblogin

import (
	"database/sql"
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
	EventMaxName             = "1234567890" // Event defined as varchar(10).
)

// Event represents a system event, such as a user login or registration.
type Event struct {
	Name     EventName // Name of the event.
	Success  bool      // Indicates if the event was successful or not.
	UserName string    // Username associated with the event.
	Message  string    // Message or details about the event.
	Created  time.Time // Timestamp when event is recorded. Read-only, set by db when inserted.
}

// LogValue implements slog.LogValuer to return a group of the Event fields.
func (e Event) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("Name", string(e.Name)),
		slog.String("Success", fmt.Sprintf("%t", e.Success)),
		slog.String("UserName", e.UserName),
		slog.String("Message", e.Message),
		slog.Time("Created", e.Created),
	)
}

var (
	ErrWriteEventInvalidDB = errors.New("invalid db connection")
	ErrWriteEventFailed    = errors.New("failed to write event")
)

// WriteEvent records an event to the database.
func WriteEvent(db *sql.DB, event Event) error {
	if db == nil {
		slog.Error("db is nil", "func", "WriteEvent")
		return ErrWriteEventInvalidDB
	}

	const qry = `INSERT INTO events(name, success, userName, message) VALUES(?, ?, ?, ?)`
	_, err := db.Exec(qry, event.Name, event.Success, event.UserName, event.Message)
	if err != nil {
		slog.Error("failed to write event", "err", err, "event", event)
		return fmt.Errorf("%w: %v", ErrWriteEventFailed, err)
	}

	slog.Debug("successfully wrote event to database", "event", event)
	return nil
}
