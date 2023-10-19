package webhandler_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

func TestAttachRequestLogger(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "GET Request",
			method:         http.MethodGet,
			url:            "/test",
			expectedStatus: http.StatusOK,
		},
		// Add more test cases as needed
	}

	handler, err := webhandler.New(webhandler.WithAppName("TestApp"))
	if err != nil {
		t.Fatalf("Error creating handler: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}

			rec := httptest.NewRecorder()

			// Create a next handler that retrieves the logger from the context
			// and writes a 200 OK status.
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				logger := webhandler.Logger(r.Context())
				if reflect.DeepEqual(logger, slog.Default()) {
					t.Error("Expected a context-specific logger, got the default logger")
				}
				w.WriteHeader(http.StatusOK)
			})

			middleware := handler.AttachRequestLogger(next)
			middleware.ServeHTTP(rec, req)

			if status := rec.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check if the logger retrieval works without a valid context
			if logger := webhandler.Logger(nil); !reflect.DeepEqual(logger, slog.Default()) {
				t.Error("Expected the default logger when passing a nil context")
			}
		})
	}
}
