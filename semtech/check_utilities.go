// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"reflect"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func checkErrors(t *testing.T, want bool, got error) {
	if (!want && got == nil) || (want && got != nil) {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Received error does not match expectation. Got: %v", got)
}

func checkHeaders(t *testing.T, want []byte, got []byte) {
	l := len(want)
	if len(got) < l {
		Ko(t, "Received header does not match expectations.\nWant: %+x\nGot:  %+x", want, got)
		return
	}
	if !reflect.DeepEqual(want[:], got[:l]) {
		Ko(t, "Received header does not match expectations.\nWant: %+x\nGot:  %+x", want, got[:l])
		return
	}
	Ok(t, "Check Headers")
}

func checkJSON(t *testing.T, want string, got []byte) {
	l := len([]byte(want))
	if len(got) < l {
		Ko(t, "Received JSON does not match expectations.\nWant: %s\nGot:  %v", want, got)
		return
	}
	str := string(got[len(got)-l:])
	if want != str {
		Ko(t, "Received JSON does not match expectations.\nWant: %s\nGot:  %s", want, str)
		return
	}
	Ok(t, "Check JSON")
}

func checkPackets(t *testing.T, want Packet, got Packet) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check packets")
		return
	}
	Ko(t, "Received packet does not match expectations.\nWant: %s\nGot:  %s", want.String(), got.String())
}
