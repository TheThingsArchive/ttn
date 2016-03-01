// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

// Checks that two packets match
func checkPackets(t *testing.T, want Packet, got Packet) {
	if want == nil {
		if got == nil {
			Ok(t, "Check packets")
			return
		}
		Ko(t, "No packet was expected but got %s", got.String())
		return
	}

	if got == nil {
		Ko(t, "Was expecting %s but got nothing", want.String())
		return
	}

	if reflect.DeepEqual(want, got) {
		Ok(t, "Check packets")
		return
	}

	Ko(t, "Converted packet does not match expectations. \nWant: \n%s\nGot:  \n%s", want.String(), got.String())
}

// Checks that obtained json matches expected one
func checkJSON(t *testing.T, want string, got []byte) {
	str := string(got)
	if str == want {
		Ok(t, "check JSON")
		return
	}
	Ko(t, "Marshaled data does not match expectations.\nWant: %s\nGot:  %s", want, str)
	return
}

// Checks that obtained metadata matches expected one
func checkMetadata(t *testing.T, want Metadata, got Metadata) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "check Metadata")
		return
	}
	Ko(t, "Unmarshaled json does not match expectations. \nWant: %s\nGot:  %s", want.String(), got.String())
}

// Check that obtained json strings contains the required field
func checkFields(t *testing.T, want []string, got []byte) {
	for _, field := range want {
		ok, err := regexp.Match(fmt.Sprintf("\"%s\":", field), got)
		if !ok || err != nil {
			Ko(t, "Expected field %s in %s", field, string(got))
			return
		}
	}
	Ok(t, "Check fields")
}
