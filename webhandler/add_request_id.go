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

// randomLower generates a random string of the specified length that
// is composed of lowercase letters.
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

// generateRequestID generates a unique request ID based on an atomic counter.
func generateRequestID(counter *uint32) string {
	// Combine the random prefix with a hexadecimal counter.
	return fmt.Sprintf("%s%08X", reqIDPrefix, atomic.AddUint32(counter, 1))
}

// reqIDType is a unique key type to avoid collisions with other context values.
type reqIDType struct{}

// reqIDKey is a key for storing and retrieving the request ID from the context.
var reqIDKey = reqIDType{}

// AddRequestID is middleware that generates a unique request ID for each
// incoming HTTP request and adds it to the request's context.
// It uses an atomic counter to ensure that each request ID is unique.
func (h Handler) AddRequestID(next http.Handler) http.Handler {
	var counter uint32

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a unique request ID using an atomic counter.
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
// If the context is nil or does not contain a request ID, an empty string
// is returned.
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
