// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"
)

// Token represent a token for the user.
type Token struct {
	Value   string
	Expires time.Time
	Kind    string
}

// Hash returns a hex encoded sha256 Hash of the given string.
// TODO: should this be a salted Hash to be more secure?
func Hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// CreateToken creates and saves a token for user of size that expires in duration.
func (db *AuthDB) CreateToken(kind, username string, size int, duration string) (Token, error) {
	var err error

	token := Token{Kind: kind}

	token.Value, err = GenerateRandomString(size)
	if err != nil {
		return Token{}, err
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		return Token{}, err
	}
	token.Expires = time.Now().Add(d)
	slog.Debug("SaveNewToken",
		"duration", duration, "d", d.String(), "expires", token.Expires.String())

	// hash the token to avoid reuse if database is compromised
	hashedValue := Hash(token.Value)

	// Insert token into database but ensure username exists.
	qry := `INSERT INTO tokens (hashedValue, expires, kind, username) SELECT ?, ?, ?, ? FROM users WHERE EXISTS (SELECT 1 FROM users WHERE username = ?) LIMIT 1`
	result, err := db.Exec(qry, hashedValue, token.Expires, kind, username, username)
	if err != nil {
		return Token{}, err
	}

	// Confirm one row was inserted.
	rows, err := result.RowsAffected()
	if err != nil {
		return Token{}, err
	}
	if rows != 1 {
		return Token{}, ErrUserNotFound
	}

	return token, err
}

var ErrTokenNotFound = errors.New("token not found")

// RemoveToken removes the token with kind and value.
func (db *AuthDB) RemoveToken(kind, value string) error {
	hashedValue := Hash(value)

	const qry = "DELETE FROM tokens WHERE kind = ? AND hashedValue = ?"
	result, err := db.Exec(qry, kind, hashedValue)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrTokenNotFound
	}

	return nil
}
