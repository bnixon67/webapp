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

// GetCookieValue returns the Value for the named cookie or an empty string if not found or an error occurs.
func GetCookieValue(r *http.Request, name string) (string, error) {
	var value string
	if r == nil {
		return value, ErrRequestNil
	}

	cookie, err := r.Cookie(name)
	if err != nil {
		// ignore ErrNoCookie
		if !errors.Is(err, http.ErrNoCookie) {
			return value, err
		}
	} else {
		value = cookie.Value
	}

	return value, nil
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
