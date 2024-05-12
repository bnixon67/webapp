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
var reqIDPrefix string

// reqIDPrefixLength is the length of the random request ID prefix.
const reqIDPrefixLength = 4

// init generates a unique request id prefix at program start.
func init() {
	const lowerLetters = "abcdefghijklmnopqrstuvwxyz"

	var err error
	reqIDPrefix, err = RandomStringFromCharset(lowerLetters, reqIDPrefixLength)
	if err != nil {
		panic("failed to initialize request ID prefix: " + err.Error())
	}
}

// generateRequestID generates a unique request ID by combining a random prefix
// with a hexadecimal representation of an incremented atomic counter. This
// ensures that each request ID is both unique and contains some randomization.
func generateRequestID(counter *uint32) string {
	return fmt.Sprintf("%s%08X", reqIDPrefix, atomic.AddUint32(counter, 1))
}

// reqIDType is a unique key type to avoid collisions with other packages.
type reqIDType struct{}

// reqIDKey is a key instance to store/retrieve request ID from the context.
var reqIDKey = reqIDType{}

// NewRequestIDMiddleware returns middleware that enhances the incoming
// HTTP request by adding a unique request ID. This ID is added both to
// the request's context and as the 'X-Request-ID' header in the response.
// The request ID is generated using an atomic counter to ensure uniqueness.
func NewRequestIDMiddleware(next http.Handler) http.Handler {
	var counter uint32 // Persist across requests

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := generateRequestID(&counter)

		ctx := context.WithValue(r.Context(), reqIDKey, reqID)

		w.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID extracts the request ID from the provided context. If the context
// is nil or does not include a request ID, the function returns an empty
// string.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if reqID, ok := ctx.Value(reqIDKey).(string); ok {
		return reqID
	}

	return ""
}
