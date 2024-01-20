// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents in the application.
type User struct {
	Username        string
	FullName        string
	Email           string
	IsAdmin         bool
	Confirmed       bool
	Created         time.Time
	LastLoginTime   time.Time
	LastLoginResult string // TODO: implement as bool?
}

// LogValue implements slog.LogValuer to group User fields in log output.
func (u User) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("Username", u.Username),
		slog.String("Fullname", u.FullName),
		slog.String("Email", u.Email),
		slog.Bool("IsAdmin", u.IsAdmin),
		slog.Bool("Confirmed", u.Confirmed),
		slog.Time("Created", u.Created),
		slog.Time("LastLoginTime", u.LastLoginTime),
		slog.String("LastLoginResult", u.LastLoginResult),
	)
}

// Define command error values.
var (
	ErrInvalidDB                 = errors.New("invalid db")
	ErrUserSessionNotFound       = errors.New("user session not found")
	ErrUserNotFound              = errors.New("user not found")
	ErrUserSessionExpired        = errors.New("user session expired")
	ErrUserAlreadyConfirmed      = errors.New("user already confirmed")
	ErrResetPasswordTokenExpired = errors.New("reset password token expired")
	ErrConfirmTokenExpired       = errors.New("confirm token expired")
	ErrUserGetLastLoginFailed    = errors.New("failed to get user last login")
)

var EmptyUser User // EmptyUser is a empty User used when returning a error.

// UserForSessionToken returns a user for the given sessionToken.
func (db *LoginDB) UserForSessionToken(sessionToken string) (User, error) {
	var (
		expires time.Time
		user    User
	)

	hashedValue := hash(sessionToken)

	if db == nil {
		return EmptyUser, ErrInvalidDB
	}

	qry := `SELECT users.userName, fullName, email, expires, admin, confirmed, users.created FROM users INNER JOIN tokens ON users.userName=tokens.userName WHERE tokens.kind = "session" AND hashedValue=? LIMIT 1`
	result := db.QueryRow(qry, hashedValue)
	err := result.Scan(&user.Username, &user.FullName, &user.Email, &expires, &user.IsAdmin, &user.Confirmed, &user.Created)
	if err != nil {
		// return custom error and empty user if session not found
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("unexpected",
				"err", ErrUserSessionNotFound,
				"sessionToken", sessionToken)
			return EmptyUser, ErrUserSessionNotFound
		}
		return EmptyUser, err
	}

	// return empty user if session is expired
	if expires.Before(time.Now()) {
		slog.Warn("unexpected",
			"err", ErrUserSessionExpired,
			"expires", expires,
			"user", user)
		return EmptyUser, ErrUserSessionExpired
	}

	user.LastLoginTime, user.LastLoginResult, err = db.LastLoginForUser(user.Username)
	if err != nil {
		return user, fmt.Errorf("%w: %v", ErrUserGetLastLoginFailed, err)
	}

	return user, err
}

// UserForName returns a user for the given userName.
func (db *LoginDB) UserForName(userName string) (User, error) {
	var user User

	qry := `SELECT userName, fullName, email, admin FROM users WHERE userName=? LIMIT 1`
	result := db.QueryRow(qry, userName)
	err := result.Scan(&user.Username, &user.FullName, &user.Email, &user.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmptyUser, ErrUserNotFound
		}
		return EmptyUser, err
	}

	return user, err
}

// UserExists returns true if the given userName already exists in db.
func (db *LoginDB) UserExists(userName string) (bool, error) {
	return db.RowExists("SELECT 1 FROM users WHERE userName=? LIMIT 1", userName)
}

// EmailExists returns true if the given email already exists.
func (db *LoginDB) EmailExists(email string) (bool, error) {
	return db.RowExists("SELECT 1 FROM users WHERE email=? LIMIT 1", email)
}

// UsernameForEmail looks up a username based on their email address.
//
// If not found, ErrUserNotFound is returned.
//
// If a SQL error occurs, other than ErrNoRows, it is returned.
func (db *LoginDB) UsernameForEmail(email string) (string, error) {
	var username string

	row := db.QueryRow("SELECT username FROM users WHERE email=?", email)
	err := row.Scan(&username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	return username, err
}

// UsernameForResetToken returns the userName for a given reset token.
func (db *LoginDB) UsernameForResetToken(tokenValue string) (string, error) {
	var userName string
	var expires time.Time
	hashedValue := hash(tokenValue)

	qry := `SELECT userName, expires FROM tokens WHERE kind="reset" AND hashedValue=?`
	row := db.QueryRow(qry, hashedValue)
	err := row.Scan(&userName, &expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	// check if token is expired
	if expires.Before(time.Now()) {
		db.RemoveToken("reset", tokenValue)
		return "", ErrResetPasswordTokenExpired
	}

	return userName, err
}

// UsernameForConfirmToken returns the userName for a given confirm token.
//
// If token is not found, ErrUserNotFound is returned.
//
// If token is expired, ErrConfirmTokenExpired is returned and token is removed.
//
// If a SQL error occurs, it will be returned, except ErrNoRows.
func (db *LoginDB) UsernameForConfirmToken(tokenValue string) (string, error) {
	var userName string
	var expires time.Time
	hashedValue := hash(tokenValue)

	qry := `SELECT tokens.userName, tokens.expires FROM tokens JOIN users ON tokens.userName = users.userName WHERE kind="confirm" AND hashedValue=? LIMIT 1`
	row := db.QueryRow(qry, hashedValue)
	err := row.Scan(&userName, &expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	// Check if token is expired.
	if expires.Before(time.Now()) {
		db.RemoveToken("confirm", tokenValue)
		return "", ErrConfirmTokenExpired
	}

	return userName, err
}

var ErrIncorrectPassword = errors.New("incorrect password")

// CompareUserPassword compares the password and hashed password for the user.
// Returns nil on success or an error on failure.
func (db *LoginDB) CompareUserPassword(userName, password string) error {
	if db == nil {
		return errors.New("invalid db")
	}

	// get hashed password for the given user
	qry := `SELECT hashedPassword FROM users WHERE username=? LIMIT 1`
	result := db.QueryRow(qry, userName)

	var hashedPassword string
	err := result.Scan(&hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	// compared hashed password with given password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrIncorrectPassword, err)
	}

	return nil
}

// RegisterUser registers a user with the given values.
// Returns nil on success or an error on failure.
func (db *LoginDB) RegisterUser(userName, fullName, email, password string) error {
	// hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// store the user and hashed password
	_, err = db.Exec("INSERT INTO users(username, hashedPassword, fullName, email) VALUES (?, ?, ?, ?)",
		userName, hashedPassword, fullName, email)
	if err != nil {
		return err
	}

	return nil
}

// LastLoginForUser retrieves the last login time and result for a given userName.  It returns zero values in case of no previous login.
func (db *LoginDB) LastLoginForUser(userName string) (time.Time, string, error) {
	var lastLogin time.Time
	var success string

	if db == nil {
		return lastLogin, success, errors.New("invalid db")
	}

	// get the second row, if it exists, since first row is current login
	qry := `SELECT created, success FROM events WHERE userName = ? AND name = ? ORDER BY created DESC LIMIT 1 OFFSET 1`
	row := db.QueryRow(qry, userName, EventLogin)
	err := row.Scan(&lastLogin, &success)
	if err != nil {
		// ignore ErrNoRows since there may not be a last login
		if errors.Is(err, sql.ErrNoRows) {
			return lastLogin, success, nil
		}
		return lastLogin, success, err
	}

	return lastLogin, success, nil
}

const SessionTokenCookieName = "session"

// UserFromRequest returns the user for the session token cookie in the request.
// If the session token is invalid or expired, the cookie is removed and an empty user returned.
func (db *LoginDB) UserFromRequest(w http.ResponseWriter, r *http.Request) (User, error) {
	// Get value of the session token cookie from the request.
	sessionToken, err := CookieValue(r, SessionTokenCookieName)
	if err != nil {
		return User{}, err
	}

	// Return an empty User struct if the session token is empty.
	if sessionToken == "" {
		return User{}, err
	}

	// Get user associated with the session token.
	user, err := db.UserForSessionToken(sessionToken)
	if err != nil {
		// Clear cookie if session is invalid or expired token.
		http.SetCookie(w, &http.Cookie{
			Name:   SessionTokenCookieName,
			Value:  "",
			MaxAge: -1,
		})

		// Ignore session not found or expired errors.
		if errors.Is(err, ErrUserSessionNotFound) || errors.Is(err, ErrUserSessionExpired) {
			return User{}, nil
		}

		return User{}, err
	}

	return user, err
}

// ConfirmUser updates database to indicate user confirmed their email.
//
// If a SQL error occurs, it will be returned.
//
// If username is not found, ErrUserNotFound is returned.
func (db *LoginDB) ConfirmUser(username string) error {
	// Check if user is already confirmed.
	alreadyConfirmed, err := db.RowExists("SELECT 1 FROM users WHERE confirmed = ? AND username = ?", true, username)
	if err != nil {
		return err
	}
	if alreadyConfirmed {
		return ErrUserAlreadyConfirmed
	}

	const qry = "UPDATE users SET confirmed = ? WHERE username = ?"
	result, err := db.Exec(qry, true, username)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrUserNotFound
	}

	return nil
}
