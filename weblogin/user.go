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

// User represents a user stored in the database.
type User struct {
	UserName        string
	FullName        string
	Email           string
	IsAdmin         bool
	Created         time.Time
	LastLoginTime   time.Time
	LastLoginResult string
}

// LogValue implements slog.LogValuer.
// It returns a group containing the fields of User so they appear together in the log output.
func (u User) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("UserName", u.UserName),
		slog.String("Fullname", u.FullName),
		slog.String("Email", u.Email),
		slog.String("IsAdmin", fmt.Sprintf("%t", u.IsAdmin)),
		slog.Time("Created", u.Created),
		slog.Time("LastLoginTime", u.LastLoginTime),
		slog.String("LastLoginResult", u.LastLoginResult),
	)
}

// Define command error values.
var (
	ErrUserSessionNotFound       = errors.New("user session not found")
	ErrUserNotFound              = errors.New("user not found")
	ErrUserSessionExpired        = errors.New("user session expired")
	ErrResetPasswordTokenExpired = errors.New("reset password token expired")
	ErrUserGetLastLoginFailed    = errors.New("user failed to get last login")
)

// GetUserForSessionToken returns a user for the given sessionToken.
func (db *LoginDB) GetUserForSessionToken(sessionToken string) (User, error) {
	var (
		expires time.Time
		user    User
	)

	hashedValue := hash(sessionToken)

	qry := `SELECT users.userName, fullName, email, expires, admin, users.created FROM users INNER JOIN tokens ON users.userName=tokens.userName WHERE tokens.kind = "session" AND hashedValue=? LIMIT 1`
	if db == nil {
		return User{}, errors.New("invalid db")
	}

	result := db.QueryRow(qry, hashedValue)
	err := result.Scan(&user.UserName, &user.FullName, &user.Email, &expires, &user.IsAdmin, &user.Created)
	if err != nil {
		// return custom error and empty user if session not found
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("unexpected",
				"err", ErrUserSessionNotFound,
				"sessionToken", sessionToken)
			return User{}, ErrUserSessionNotFound
		}
		return User{}, err
	}

	// return empty user if session is expired
	if expires.Before(time.Now()) {
		slog.Warn("unexpected",
			"err", ErrUserSessionExpired,
			"expires", expires,
			"user", user)
		return User{}, ErrUserSessionExpired
	}

	user.LastLoginTime, user.LastLoginResult, err = db.LastLoginForUser(user.UserName)
	if err != nil {
		return user, fmt.Errorf("%w: %v", ErrUserGetLastLoginFailed, err)
	}

	return user, err
}

// GetUserForName returns a user for the given userName.
func (db *LoginDB) GetUserForName(userName string) (User, error) {
	var user User

	qry := `SELECT userName, fullName, email, admin FROM users WHERE userName=? LIMIT 1`
	result := db.QueryRow(qry, userName)
	err := result.Scan(&user.UserName, &user.FullName, &user.Email, &user.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
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

// GetUserNameForEmail returns the userName for a given email.
func (db *LoginDB) GetUserNameForEmail(email string) (string, error) {
	var userName string

	row := db.QueryRow("SELECT username FROM users WHERE email=?", email)
	err := row.Scan(&userName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	return userName, err
}

// GetUserNameForResetToken returns the userName for a given reset token.
func (db *LoginDB) GetUserNameForResetToken(tokenValue string) (string, error) {
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

// GetUserFromRequest returns the current User or empty User if the session is not found.
func (db *LoginDB) GetUserFromRequest(w http.ResponseWriter, r *http.Request) (User, error) {
	var user User

	// get sessionToken from cookie, if it exists
	sessionToken, err := GetCookieValue(r, SessionTokenCookieName)
	if err != nil {
		return user, err
	}

	// get user if there is a sessionToken
	if sessionToken != "" {
		user, err = db.GetUserForSessionToken(sessionToken)
		if err != nil {
			// delete invalid token to prevent session fixation
			http.SetCookie(w,
				&http.Cookie{
					Name:   SessionTokenCookieName,
					Value:  "",
					MaxAge: -1,
				})
		}
		// ignore session not found or expired errors
		if errors.Is(err, ErrUserSessionNotFound) || errors.Is(err, ErrUserSessionExpired) {
			err = nil
		}
	}

	return user, err
}
