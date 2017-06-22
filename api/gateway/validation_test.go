// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"testing"

	s "github.com/smartystreets/assertions"
)

func TestValidation(t *testing.T) {
	table := []struct {
		subject  *LocationMetadata
		expected error
	}{
		{&LocationMetadata{0, 0, 0, 0, 0}, ErrLocationZero},
		{&LocationMetadata{0, -0.001, 0.001, 0, 0}, ErrLocationZero},
		{&LocationMetadata{0, 300, 0, 0, 0}, ErrInvalidLatitude},
		{&LocationMetadata{0, -300, 0, 0, 0}, ErrInvalidLatitude},
		{&LocationMetadata{0, 0, 300, 0, 0}, ErrInvalidLongitude},
		{&LocationMetadata{0, 0, -300, 0, 0}, ErrInvalidLongitude},
		{&LocationMetadata{0, 12, 34, 10, 0}, nil},
	}
	for i, tt := range table {
		t.Run(fmt.Sprintf("Table %d", i), func(t *testing.T) {
			a := s.New(t)
			a.So(tt.subject.Validate(), s.ShouldEqual, tt.expected)
		})
	}
}
