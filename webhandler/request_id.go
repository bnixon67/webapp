// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
)

// reqIDPrefix is a random prefix for the request ID set at program startup.
var reqIDPrefix string = generateRandomPrefix()

// reqIDPrefixLength is the length of the random request ID prefix.
const reqIDPrefixLength = 4

// generateRandomPrefix creates a random string to be used as a prefix for
// generating request IDs.
//
// If the random string generation fails, the function will panic.
func generateRandomPrefix() string {
	const lowerLetters = "abcdefghijklmnopqrstuvwxyz"

	prefix, err := RandomStringFromCharset(lowerLetters, reqIDPrefixLength)
	if err != nil {
		panic("failed to initialize request ID prefix: " + err.Error())
	}

	return prefix
}

// generateRequestID generates a unique request ID by concatenating a
// pre-defined random prefix with the hexadecimal representation of an
// atomically incremented counter.
func generateRequestID(counter *uint32) string {
	id := atomic.AddUint32(counter, 1)
	return fmt.Sprintf("%s%08X", reqIDPrefix, id)
}

// reqIDType is a custom type to avoid collisions in context values.
type reqIDType struct{}

// reqIDKey is a unique identifier to store/retrieve request ID from a context.
var reqIDKey = reqIDType{}

// NewRequestIDMiddleware creates middleware that assigns a unique request
// ID to every incoming HTTP request. This ID is added to the request's
// context and set as the 'X-Request-ID' header in the HTTP response.
//
// It uses an atomic counter to ensure each ID is unique across all requests.
func NewRequestIDMiddleware(next http.Handler) http.Handler {
	var counter uint32 // Counter to generate unique IDs, persistent across requests.

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := generateRequestID(&counter)

		w.Header().Set("X-Request-ID", reqID)

		ctx := context.WithValue(r.Context(), reqIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID extracts the request ID from the provided context.
//
// If the context is nil or does not include a request ID, the function
// returns an empty string.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if reqID, ok := ctx.Value(reqIDKey).(string); ok {
		return reqID
	}

	return ""
}
