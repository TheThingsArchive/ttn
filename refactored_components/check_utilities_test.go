// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/thethingsnetwork/core"
	. "github.com/thethingsnetwork/core/utils/testing"
	"reflect"
	"testing"
)

// Checks that two core packets match
func checkPackets(t *testing.T, want core.Packet, got core.Packet) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check packets")
		return
	}
	Ko(t, "Converted packet don't match expectations. \nWant: \n%s\nGot:  \n%s", want.String(), got.String())
}

// Checks that errors match
func checkErrors(t *testing.T, want error, got error) {
	if got == want {
		Ok(t, "check Errors")
		return
	}
	Ko(t, "Expected error to be %v but got %v", want, got)
}

// Checks that obtained json matches expected one
func checkJSON(t *testing.T, want string, got []byte) {
	str := string(got)
	if str == want {
		Ok(t, "check JSON")
		return
	}
	Ko(t, "Marshaled data don't match expectations.\nWant: %s\nGot:  %s", want, str)
	return
}

// Checks that obtained metadata matches expected one
func checkMetadata(t *testing.T, want Metadata, got Metadata) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "check Metadata")
		return
	}
	Ko(t, "Unmarshaled json don't match expectations. \nWant: %s\nGot:  %s", want.String(), got.String())
}
