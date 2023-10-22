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

// reqIDType is a unique key type to avoid collisions with other context values.
type reqIDType struct{}

// reqIDKey is a key for storing and retrieving the request ID from the context.
var reqIDKey = reqIDType{}

// reqIDPrefix is a random prefix for the request ID set at program startup.
var reqIDPrefix string

// randomLower generates a random string of the specified length, composed of
// lowercase letters.
func randomLower(length int) (string, error) {
	const lower = "abcdefghijklmnopqrstuvwxyz"

	// check for valid length
	if length <= 0 {
		return "", errors.New("invalid length")
	}

	result := make([]byte, length)

	for i := 0; i < length; i++ {
		// generate random index
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(lower))))
		if err != nil {
			return "", err
		}
		result[i] = lower[idx.Int64()]
	}

	return string(result), nil
}

// init generates a unique request id prefix at program start.
func init() {
	reqIDPrefix, _ = randomLower(4)
}

// AddRequestID is middleware that generates a unique request ID for each
// incoming HTTP request and adds it to the request's context.
// It uses an atomic counter to ensure that each request ID is unique.
func (h Handler) AddRequestID(next http.Handler) http.Handler {
	var counter uint32

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a unique request ID using an atomic counter.
		reqID := fmt.Sprintf("%s%08X",
			reqIDPrefix, atomic.AddUint32(&counter, 1))

		// Add the request ID to the request's context.
		ctx := context.WithValue(r.Context(), reqIDKey, reqID)

		// Call the next handler in the chain with the updated context.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID retrieves the request ID from the context.
// If the context is nil or does not contain a request ID, zero is returned.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	reqID, ok := ctx.Value(reqIDKey).(string)
	if !ok {
		return ""
	}

	return reqID
}
