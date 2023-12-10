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
func SliceOfStructsToCSV(w io.Writer, data interface{}) error {
	sliceValue := reflect.ValueOf(data)
	if sliceValue.Kind() != reflect.Slice {
		return ErrCSVNotSlice
	}

	cw := csv.NewWriter(w)
	// Don't defer cw.Flush() given the buffered behavior of csv.Writer

	// Write CSV header
	if sliceValue.Len() > 0 {
		firstElem := sliceValue.Index(0)
		if firstElem.Kind() != reflect.Struct {
			return ErrCSVNotSliceOfStructs
		}

		var header []string
		t := firstElem.Type()
		for i := 0; i < firstElem.NumField(); i++ {
			header = append(header, t.Field(i).Name)
		}

		if err := cw.Write(header); err != nil {
			return err
		}
	}

	// Write CSV rows
	for i := 0; i < sliceValue.Len(); i++ {
		structValue := sliceValue.Index(i)
		var record []string

		for j := 0; j < structValue.NumField(); j++ {
			field := structValue.Field(j)
			record = append(record, fmt.Sprintf("%v", field.Interface()))
		}

		if err := cw.Write(record); err != nil {
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
