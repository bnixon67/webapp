// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
)

var (
	ErrCSVNotSlice          = errors.New("data is not a slice")
	ErrCSVNotSliceOfStructs = errors.New("slice elements are not structs")
	ErrCSVWriteFailed       = errors.New("failed to write")
)

// SliceOfStructsToCSV writes a slice of structs to a CSV writer using struct
// field names as headers.  This function assumes that the data provided is
// a slice of structs.
func SliceOfStructsToCSV(w io.Writer, data interface{}) error {
	sliceValue := reflect.ValueOf(data)
	if sliceValue.Kind() != reflect.Slice {
		return ErrCSVNotSlice
	}

	cw := csv.NewWriter(w)
	// Don't defer cw.Flush() to allow checking for Errors

	for i := 0; i < sliceValue.Len(); i++ {
		structValue := sliceValue.Index(i)

		if structValue.Kind() != reflect.Struct {
			return ErrCSVNotSliceOfStructs
		}

		// write headers if first row
		if i == 0 {
			if err := writeHeader(cw, structValue); err != nil {
				return err
			}
		}

		if err := writeRecord(cw, structValue); err != nil {
			return err
		}
	}

	cw.Flush() // Call Flush explicitly to ensure all data is written.

	// Check for any errors that occurred during the write operations
	if err := cw.Error(); err != nil {
		return fmt.Errorf("%w: %v", ErrCSVWriteFailed, err)
	}

	return nil
}

// writeHeader writes CSV headers from the struct field names or `csv` tags.
func writeHeader(cw *csv.Writer, v reflect.Value) error {
	var headers []string
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		name := field.Tag.Get("csv")
		if name == "" {
			name = field.Name
		}
		headers = append(headers, name)
	}
	return cw.Write(headers)
}

// writeRecord converts struct fields to CSV record.
func writeRecord(cw *csv.Writer, v reflect.Value) error {
	var records []string
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		records = append(records, fmt.Sprintf("%v", field.Interface()))
	}
	return cw.Write(records)
}
