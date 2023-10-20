package webhandler

import (
	"context"
	"net/http"
	"sync/atomic"
)

type ctxKey int

const reqIDKey ctxKey = iota

var reqCounter int64

func (h Handler) AddRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Generate a unique request ID using an atomic counter.
		reqID := atomic.AddInt64(&reqCounter, 1)

		// Add the logger to the request's context.
		ctx := context.WithValue(r.Context(), reqIDKey, reqID)

		// Call the next handler in the chain with the updated context.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID retrieves the request ID from the context.
func RequestID(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}

	reqID, ok := ctx.Value(reqIDKey).(int64)
	if !ok {
		return 0
	}

	return reqID
}
