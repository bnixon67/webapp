// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp_test

import (
	"testing"

	"github.com/bnixon67/webapp/webapp"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		opts     []webapp.Option
		wantName string
		wantErr  bool
	}{
		{
			name:     "With AppName",
			opts:     []webapp.Option{webapp.WithName("TestApp")},
			wantName: "TestApp",
			wantErr:  false,
		},
		{
			name:     "Without AppName",
			opts:     []webapp.Option{},
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := webapp.New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && h.Name != tt.wantName {
				t.Errorf("New() AppName = %v, want %v", h.Name, tt.wantName)
			}
		})
	}
}
