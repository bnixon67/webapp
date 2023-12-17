package webutil_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/bnixon67/webapp/webutil"
)

func TestTemplates(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		wantErr   bool
		wantTmpls []string
	}{
		{
			name:    "InvalidPattern",
			pattern: "nonexistent/*.tmpl",
			wantErr: true,
		},
		{
			name:      "ValidPattern",
			pattern:   "testdata/*.tmpl",
			wantErr:   false,
			wantTmpls: []string{"tmpl", "tmpl1.tmpl"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := webutil.Templates(tc.pattern)
			if (err != nil) != tc.wantErr {
				t.Errorf("Templates(%q) error = %v, wantErr %v",
					tc.pattern, err, tc.wantErr)
				return
			}
			if err == nil {
				tmplNames := webutil.TemplateNames(got)
				slices.Sort(tmplNames)
				slices.Sort(tc.wantTmpls)

				if !reflect.DeepEqual(tmplNames, tc.wantTmpls) {
					t.Errorf("Templates(%q) got = %v, want %v", tc.pattern, tmplNames, tc.wantTmpls)
				}
			}
		})
	}
}
