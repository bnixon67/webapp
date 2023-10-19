package webhandler_test

import (
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

// TestNew is a table-driven test that checks the New function.
func TestNew(t *testing.T) {
	tests := []struct {
		name     string              // name is the name of the test case
		opts     []webhandler.Option // opts are the options to pass to the New function
		wantName string              // wantName is the expected AppName of the created Handler
		wantErr  bool                // wantErr indicates whether an error is expected
	}{
		{
			name:     "With AppName",
			opts:     []webhandler.Option{webhandler.WithAppName("TestApp")},
			wantName: "TestApp",
			wantErr:  false,
		},
		{
			name:     "Without AppName",
			opts:     []webhandler.Option{},
			wantName: "",
			wantErr:  true,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := webhandler.New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && h.AppName != tt.wantName {
				t.Errorf("New() AppName = %v, want %v", h.AppName, tt.wantName)
			}
		})
	}
}
