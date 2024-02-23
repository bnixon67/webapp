// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

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
	ErrUserLoginTokenNotFound    = errors.New("user login token not found")
	ErrUserNotFound              = errors.New("user not found")
	ErrUserLoginTokenExpired     = errors.New("user login expired")
	ErrUserAlreadyConfirmed      = errors.New("user already confirmed")
	ErrResetPasswordTokenExpired = errors.New("reset password token expired")
	ErrConfirmTokenExpired       = errors.New("confirm token expired")
	ErrUserGetLastLoginFailed    = errors.New("failed to get user last login")
)

var EmptyUser User // EmptyUser is a empty User used when returning a error.

// UserForLoginToken returns a user for the given loginToken.
func (db *LoginDB) UserForLoginToken(loginToken string) (User, error) {
	var (
		expires time.Time
		user    User
	)

	hashedValue := Hash(loginToken)

	if db == nil {
		return EmptyUser, ErrInvalidDB
	}

	qry := `SELECT users.username, fullName, email, expires, admin, confirmed, users.created FROM users INNER JOIN tokens ON users.username=tokens.username WHERE tokens.kind = ? AND hashedValue=? LIMIT 1`
	result := db.QueryRow(qry, LoginTokenKind, hashedValue)
	err := result.Scan(&user.Username, &user.FullName, &user.Email, &expires, &user.IsAdmin, &user.Confirmed, &user.Created)
	if err != nil {
		// Return custom error if login not found
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("unexpected",
				"err", ErrUserLoginTokenNotFound,
				"loginToken", loginToken)
			return EmptyUser, ErrUserLoginTokenNotFound
		}
		return EmptyUser, err
	}

	// Check if login token is expired.
	if expires.Before(time.Now()) {
		slog.Warn("unexpected",
			slog.Any("err", ErrUserLoginTokenExpired),
			slog.Time("expires", expires),
			slog.Any("user", user))

		// Remove expired token.
		err := db.RemoveToken(LoginTokenKind, loginToken)
		if err != nil {
			slog.Error("failed to remove login token",
				"loginToken", loginToken, "err", err)
		}

		return EmptyUser, ErrUserLoginTokenExpired
	}

	user.LastLoginTime, user.LastLoginResult, err = db.LastLoginForUser(user.Username)
	if err != nil {
		return user, fmt.Errorf("%w: %v", ErrUserGetLastLoginFailed, err)
	}

	return user, err
}

// UserForName returns a user for the given username.
func (db *LoginDB) UserForName(username string) (User, error) {
	var user User

	qry := `SELECT username, fullName, email, admin FROM users WHERE username=? LIMIT 1`
	result := db.QueryRow(qry, username)
	err := result.Scan(&user.Username, &user.FullName, &user.Email, &user.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmptyUser, ErrUserNotFound
		}
		return EmptyUser, err
	}

	return user, err
}

// UserExists returns true if the given username already exists in db.
func (db *LoginDB) UserExists(username string) (bool, error) {
	return db.RowExists("SELECT 1 FROM users WHERE username=? LIMIT 1", username)
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

// UsernameForResetToken returns the username for a given reset token.
func (db *LoginDB) UsernameForResetToken(tokenValue string) (string, error) {
	var username string
	var expires time.Time
	hashedValue := Hash(tokenValue)

	qry := `SELECT username, expires FROM tokens WHERE kind="reset" AND hashedValue=?`
	row := db.QueryRow(qry, hashedValue)
	err := row.Scan(&username, &expires)
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

	return username, err
}

// UsernameForConfirmToken returns the username for a given confirm token.
//
// If token is not found, ErrUserNotFound is returned.
//
// If token is expired, ErrConfirmTokenExpired is returned and token is removed.
//
// If a SQL error occurs, it will be returned, except ErrNoRows.
func (db *LoginDB) UsernameForConfirmToken(tokenValue string) (string, error) {
	var username string
	var expires time.Time
	hashedValue := Hash(tokenValue)

	qry := `SELECT tokens.username, tokens.expires FROM tokens JOIN users ON tokens.username = users.username WHERE kind="confirm" AND hashedValue=? LIMIT 1`
	row := db.QueryRow(qry, hashedValue)
	err := row.Scan(&username, &expires)
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

	return username, err
}

var ErrIncorrectPassword = errors.New("incorrect password")

// AuthenticateUser compares the password and hashed password for the user.
// Returns nil on success or an error on failure.
func (db *LoginDB) AuthenticateUser(username, password string) error {
	if db == nil {
		return errors.New("invalid db")
	}

	// get hashed password for the given user
	qry := `SELECT hashedPassword FROM users WHERE username=? LIMIT 1`
	result := db.QueryRow(qry, username)

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
func (db *LoginDB) RegisterUser(username, fullName, email, password string) error {
	// hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// store the user and hashed password
	_, err = db.Exec("INSERT INTO users(username, hashedPassword, fullName, email) VALUES (?, ?, ?, ?)",
		username, hashedPassword, fullName, email)
	if err != nil {
		return err
	}

	return nil
}

// LastLoginForUser retrieves the last login time and result for a given username.  It returns zero values in case of no previous login.
func (db *LoginDB) LastLoginForUser(username string) (time.Time, string, error) {
	var lastLogin time.Time
	var success string

	if db == nil {
		return lastLogin, success, errors.New("invalid db")
	}

	// get the second row, if it exists, since first row is current login
	qry := `SELECT created, succeeded FROM events WHERE username = ? AND name = ? ORDER BY created DESC LIMIT 1 OFFSET 1`
	row := db.QueryRow(qry, username, EventLogin)
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

const LoginTokenCookieName = "login"

// UserFromRequest returns the user for the login token cookie in the request.
// If the login token is invalid or expired, the cookie is removed and
// an empty user returned.
func (db *LoginDB) UserFromRequest(w http.ResponseWriter, r *http.Request) (User, error) {
	// Get value of the login token cookie from the request.
	loginToken, err := CookieValue(r, LoginTokenCookieName)
	if err != nil {
		return User{}, err
	}

	// Return an empty User struct if the login token is empty.
	if loginToken == "" {
		return User{}, err
	}

	// Get user associated with the login token.
	user, err := db.UserForLoginToken(loginToken)
	if err != nil {
		// Clear cookie if login is invalid or expired token.
		http.SetCookie(w, &http.Cookie{
			Name:   LoginTokenCookieName,
			Value:  "",
			MaxAge: -1,
		})

		// Ignore login not found or expired errors.
		if errors.Is(err, ErrUserLoginTokenNotFound) || errors.Is(err, ErrUserLoginTokenExpired) {
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
