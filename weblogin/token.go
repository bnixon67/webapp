/*
Copyright 2023 Bill Nixon

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License.  You may obtain a copy of the
License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied.  See the License for the
specific language governing permissions and limitations under the License.
*/
package weblogin

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log/slog"
	"time"
)

// Token represent a token for the user.
type Token struct {
	Value   string
	Expires time.Time
	Kind    string
}

// hash returns a hex encoded sha256 hash of the given string.
// TODO: should this be a salted hash to be more secure?
func hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// SaveNewToken creates and saves a token for user of size that expires in duration.
func SaveNewToken(db *sql.DB, kind, userName string, size int, duration string) (Token, error) {
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
	hashedValue := hash(token.Value)

	qry := `INSERT INTO tokens(hashedValue, expires, kind, userName) VALUES(?, ?, ?, ?)`
	_, err = db.Exec(qry, hashedValue, token.Expires, kind, userName)
	return token, err
}

// RemoveToken removes the given sessionToken.
func RemoveToken(db *sql.DB, kind, value string) error {
	hashedValue := hash(value)

	qry := `DELETE FROM tokens WHERE kind = ? AND hashedValue = ?`
	_, err := db.Exec(qry, kind, hashedValue)
	return err
}
