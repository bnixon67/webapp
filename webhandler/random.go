// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"crypto/rand"
	"errors"
	"math/big"
	"unicode/utf8"
)

// RandomStringFromCharset generates a random string of specified length
// from characters in the charset. This version is Unicode-aware and works
// correctly with multi-byte characters.
func RandomStringFromCharset(charset string, length int) (string, error) {
	if utf8.RuneCountInString(charset) == 0 {
		return "", errors.New("empty or invalid charset")
	}
	if length <= 0 {
		return "", errors.New("invalid length")
	}

	runes := []rune(charset)
	charsetLen := big.NewInt(int64(len(runes)))

	buffer := make([]rune, length)
	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		buffer[i] = runes[index.Int64()]
	}

	return string(buffer), nil
}
