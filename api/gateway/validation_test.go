// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"testing"

	s "github.com/smartystreets/assertions"
)

func buildLocation(latitude, longitude float32, altitude int32) *LocationMetadata {
	return &LocationMetadata{
		Latitude:  latitude,
		Longitude: longitude,
		Altitude:  altitude,
	}
}

func TestValidation(t *testing.T) {
	table := []struct {
		subject  *LocationMetadata
		expected error
	}{
		{buildLocation(0, 0, 0), ErrLocationZero},
		{buildLocation(-0.001, 0.001, 0), ErrLocationZero},
		{buildLocation(300, 0, 0), ErrInvalidLatitude},
		{buildLocation(-300, 0, 0), ErrInvalidLatitude},
		{buildLocation(0, 300, 0), ErrInvalidLongitude},
		{buildLocation(0, -300, 0), ErrInvalidLongitude},
		{buildLocation(12, 34, 10), nil},
	}
	for i, tt := range table {
		t.Run(fmt.Sprintf("Table %d", i), func(t *testing.T) {
			a := s.New(t)
			a.So(tt.subject.Validate(), s.ShouldEqual, tt.expected)
		})
	}
}
