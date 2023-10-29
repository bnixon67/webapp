// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync/atomic"
)

// randomLower generates a random string of the specified length composed of lowercase letters.
func randomLower(length int) (string, error) {
	const lowerLetters = "abcdefghijklmnopqrstuvwxyz"

	// check for valid length
	if length <= 0 {
		return "", errors.New("invalid length")
	}

	// Create a random source using a secure random number generator.
	source := rand.Reader

	// Define the set of characters to choose from.
	charsetLength := len(lowerLetters)

	// Define location to store the result.
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		// Generate a random index within the character set.
		idx, err := rand.Int(source, big.NewInt(int64(charsetLength)))
		if err != nil {
			return "", err
		}

		// Use the random index to select a character from the set.
		result[i] = lowerLetters[idx.Int64()]
	}

	return string(result), nil
}

// reqIDPrefix is a random prefix for the request ID set at program startup.
var reqIDPrefix string

// reqIDPrefixLength is the length of the random request ID prefix.
const reqIDPrefixLength = 4

// init generates a unique request id prefix at program start.
func init() {
	var err error

	// Generate a random lowercase prefix for request IDs.
	reqIDPrefix, err = randomLower(reqIDPrefixLength)
	if err != nil {
		panic("Failed to generate request ID prefix: " + err.Error())
	}
}

// generateRequestID generates a unique request ID by combining a random prefix with a hexadecimal representation of an incremented atomic counter. This ensures that each request ID is both unique and contains some randomization.
func generateRequestID(counter *uint32) string {
	return fmt.Sprintf("%s%08X", reqIDPrefix, atomic.AddUint32(counter, 1))
}

// reqIDType is a unique key type to avoid collisions with other context values.
type reqIDType struct{}

// reqIDKey is a key for storing and retrieving the request ID from the context.
var reqIDKey = reqIDType{}

// AddRequestID is middleware that adds a unique request ID for each request to the request's context and a X-Request-ID header to the response.
func (h Handler) AddRequestID(next http.Handler) http.Handler {
	var counter uint32

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a unique request ID using a counter.
		reqID := generateRequestID(&counter)

		// Add the request ID to the request's context.
		ctx := context.WithValue(r.Context(), reqIDKey, reqID)

		// Add the request ID to headers.
		w.Header().Set("X-Request-ID", reqID)

		// Call the next handler in the chain with the updated context.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID retrieves the request ID from the context.
// If the context is nil or does not contain a request ID, an empty string is returned.
func RequestID(ctx context.Context) string {
	// Return an empty string if the context is nil.
	if ctx == nil {
		return ""
	}

	// Attempt to retrieve the request ID from the context.
	reqID, ok := ctx.Value(reqIDKey).(string)

	// If the request ID is not found in the context, return an empty string.
	if !ok {
		return ""
	}

	// Return the request ID from the context.
	return reqID
}
