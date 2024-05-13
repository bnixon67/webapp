// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil_test

import (
	"html/template"
	"testing"
	"time"

	"github.com/bnixon67/webapp/webutil"
)

func TestToTimeZone(t *testing.T) {
	// Setup a base time for testing
	baseTime := time.Date(2023, time.April, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		time    time.Time
		tzName  string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "UTCtoPST",
			time:    baseTime,
			tzName:  "America/Los_Angeles",
			want:    baseTime.In(time.FixedZone("PST", -8*3600)),
			wantErr: false,
		},
		{
			name:    "UTCtoIST",
			time:    baseTime,
			tzName:  "Asia/Kolkata",
			want:    baseTime.In(time.FixedZone("IST", 5*3600+1800)),
			wantErr: false,
		},
		{
			name:    "invalidTimezone",
			time:    baseTime,
			tzName:  "Mars/Phobos",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := webutil.ToTimeZone(tc.time, tc.tzName)

			if gotErr == nil && tc.wantErr {
				t.Errorf("Got no error but wanted one")
			}
			if gotErr != nil && !tc.wantErr {
				t.Errorf("Got error %q but wanted none", gotErr)
			}

			if !got.Equal(tc.want) {
				t.Errorf("ToTimeZone() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name  string
		elems []string
		sep   string
		want  template.HTML
	}{
		{
			name:  "joinWithSpace",
			elems: []string{"hello", "world"},
			sep:   " ",
			want:  template.HTML("hello world"),
		},
		{
			name:  "joinWithComma",
			elems: []string{"apple", "banana", "cherry"},
			sep:   ", ",
			want:  template.HTML("apple, banana, cherry"),
		},
		{
			name:  "emptyElements",
			elems: []string{"", "", ""},
			sep:   ",",
			want:  template.HTML(",,"),
		},
		{
			name:  "noSeparator",
			elems: []string{"one", "two", "three"},
			sep:   "",
			want:  template.HTML("onetwothree"),
		},
		{
			name:  "joinHTML",
			elems: []string{"one", "two", "three"},
			sep:   "<br>",
			want:  template.HTML("one<br>two<br>three"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := webutil.Join(tc.elems, tc.sep)
			if got != tc.want {
				t.Errorf("Join(%v, %q) = %q, want %q", tc.elems, tc.sep, got, tc.want)
			}
		})
	}
}
