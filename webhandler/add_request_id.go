package webhandler

import (
	"context"
	"net/http"
	"sync/atomic"
)

// reqKeyType is a unique key type to avoid key collisions with other context values.
type reqIDType struct{}

// reqIDKey is as a key for storing and retrieving the request ID from the context.
var reqIDKey = reqIDType{}

// reqCounter is a counter for generating unique request IDs.
var reqCounter uint32

// AddRequestID is middleware that generates a unique request ID for each
// incoming HTTP request and adds it to the request's context.
// It uses an atomic counter to ensure that each request ID is unique.
func (h Handler) AddRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Generate a unique request ID using an atomic counter.
		reqID := atomic.AddUint32(&reqCounter, 1)

		// Add the request ID to the request's context.
		ctx := context.WithValue(r.Context(), reqIDKey, reqID)

		// Call the next handler in the chain with the updated context.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID retrieves the request ID from the context.
// If the context is nil or does not contain a request ID, zero is returned.
func RequestID(ctx context.Context) uint32 {
	if ctx == nil {
		return 0
	}

	reqID, ok := ctx.Value(reqIDKey).(uint32)
	if !ok {
		return 0
	}

	return reqID
}
