// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package brk_hdl_http

import (
	. "github.com/thethingsnetwork/core/utils/testing"
	"testing"
)

func TestNewAdapter(t *testing.T) {
	tests := []struct {
		Port      uint
		WantError error
	}{
		{3000, nil},
		{0, ErrInvalidPort},
	}

	for _, test := range tests {
		_, err := NewAdapter(test.Port)
		checkErrors(t, test.WantError, err)
	}
}

func checkErrors(t *testing.T, want error, got error) {
	if want == got {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be {%v} but got {%v}", want, got)
}
