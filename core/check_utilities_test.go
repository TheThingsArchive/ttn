// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

// Checks that two packets match
func checkPackets(t *testing.T, want Packet, got Packet) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check packets")
		return
	}
	Ko(t, "Converted packet does not match expectations. \nWant: \n%s\nGot:  \n%s", want.String(), got.String())
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

// Checks that obtained TXPK matches expeceted one
func checkTXPKs(t *testing.T, want semtech.TXPK, got semtech.TXPK) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "check TXPKs")
		return
	}
	Ko(t, "Converted TXPK does not match expectations. \nWant: %s\nGot:  %s", pointer.DumpPStruct(want, false), pointer.DumpPStruct(got, false))
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
