// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil

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

// SliceOfStructsToCSV writes a slice of structs to a CSV writer.
// It uses the struct field names as headers and their values as rows.
func SliceOfStructsToCSV(w io.Writer, data interface{}) error {
	sliceValue := reflect.ValueOf(data)
	if sliceValue.Kind() != reflect.Slice {
		return ErrCSVNotSlice
	}

	cw := csv.NewWriter(w)
	// Don't defer cw.Flush() given the buffered behavior of csv.Writer

	// Write CSV header and rows
	for i := 0; i < sliceValue.Len(); i++ {
		structValue := sliceValue.Index(i)
		if i == 0 { // first row
			if structValue.Kind() != reflect.Struct {
				return ErrCSVNotSliceOfStructs
			}
			if err := writeHeader(cw, structValue); err != nil {
				return err
			}
		}

		err := writeRecord(cw, structValue)
		if err != nil {
			return err
		}
	}

	cw.Flush()

	// Check for any errors that occurred during the write operations
	if err := cw.Error(); err != nil {
		return fmt.Errorf("%w: %v", ErrCSVWriteFailed, err)
	}

	return nil
}

// writeHeader writes the CSV header based on struct field names.
func writeHeader(cw *csv.Writer, v reflect.Value) error {
	var header []string
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		tag := t.Field(i).Tag.Get("csv")
		if tag != "" {
			fieldName = tag
		}
		header = append(header, fieldName)
	}
	return cw.Write(header)
}

// writeRecord writes a single CSV record from a struct.
func writeRecord(cw *csv.Writer, v reflect.Value) error {
	var record []string
	for j := 0; j < v.NumField(); j++ {
		field := v.Field(j)
		record = append(record, fmt.Sprintf("%v", field.Interface()))
	}
	return cw.Write(record)
}
