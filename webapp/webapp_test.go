// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp_test

import (
	"fmt"
	"testing"

	"github.com/bnixon67/webapp/webapp"
	"github.com/google/go-cmp/cmp"
)

func TestWebAppString(t *testing.T) {
	emptyWebApp := &webapp.WebApp{}

	nameWebApp := &webapp.WebApp{}
	nameWebApp.Name = "name"

	tests := []struct {
		name   string
		webapp *webapp.WebApp
		want   string
	}{
		{
			name:   "nil webapp",
			webapp: nil,
			want:   fmt.Sprintf("%v", nil),
		},
		{
			name:   "empty webapp",
			webapp: emptyWebApp,
			want:   fmt.Sprintf("%+v", emptyWebApp),
		},
		{
			name:   "name webapp",
			webapp: nameWebApp,
			want:   fmt.Sprintf("%+v", nameWebApp),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.webapp.String()
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Errorf("mismatch for %q (-want +got):\n%s", tc.webapp, diff)

			}
		})
	}
}

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
