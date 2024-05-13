// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil

import (
	"html/template"
	"strings"
	"time"
)

// ToTimeZone returns time adjusted to the given timezone.
// If tzName is invalid, then the zero time value is returned.
func ToTimeZone(t time.Time, tzName string) (time.Time, error) {
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(loc), nil
}

// Join concatenates the elements of a []string into a single string,
// separated by sep.  The result is returned as template.HTML, which should
// not include user-controlled input without escaping.
func Join(elems []string, sep string) template.HTML {
	return template.HTML(strings.Join(elems, sep))
}
