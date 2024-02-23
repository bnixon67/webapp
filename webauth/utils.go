// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
)

var ErrInvalidLength = errors.New("invalid length")

// GenerateRandomString returns a URL safe base64 encoded string of n random bytes.
func GenerateRandomString(n int) (string, error) {
	if n < 0 {
		return "", ErrInvalidLength
	}

	// buffer to store n bytes
	b := make([]byte, n)

	// get b random bytes
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// convert to URL safe base64 encoded string
	return base64.URLEncoding.EncodeToString(b), err
}

var ErrRequestNil = errors.New("request is nil")

// CookieValue returns the named cookie value provided in the request or an empty string if not found.
func CookieValue(r *http.Request, name string) (string, error) {
	if r == nil {
		return "", ErrRequestNil
	}

	cookie, err := r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", nil // Ignore ErrNoCookie.
		}
		return "", err // Return other errors.
	}

	return cookie.Value, nil
}

// IsEmpty returns true if any of the strings are empty, otherwise false.
func IsEmpty(strs ...string) bool {
	for _, s := range strs {
		if s == "" {
			return true
		}
	}

	return false
}
