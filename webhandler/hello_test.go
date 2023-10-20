package webhandler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

func TestHelloHandler(t *testing.T) {
	tests := []struct {
		name       string // name of the test case
		method     string // HTTP method for the request
		wantStatus int    // expected HTTP status code of the response
		wantBody   string // expected body of the response
	}{
		{
			name:       "Valid GET Request",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantBody:   "hello from TestApp",
		},
		{
			name:       "Invalid POST Request",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "",
		},
		// Add more test cases as needed
	}

	handler, err := webhandler.New(webhandler.WithAppName("TestApp"))
	if err != nil {
		t.Fatalf("Error creating handler: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the specified HTTP method.
			req, err := http.NewRequest(tt.method, "/hello", nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			// Record the HTTP response.
			rec := httptest.NewRecorder()

			// Call HelloHandler.
			handler.HelloHandler(rec, req)

			// Check the status code and response body.
			if status := rec.Code; status != tt.wantStatus {
				t.Errorf("got status %v, want %v", status, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				if body := strings.TrimSpace(rec.Body.String()); body != tt.wantBody {
					t.Errorf("got body %q, want %q", body, tt.wantBody)
				}
			}
		})
	}
}
