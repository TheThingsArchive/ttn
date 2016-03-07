// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"reflect"
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

// Check utilities
func checkRegistrations(t *testing.T, want []testRegistration, got []core.RRegistration) {
	if len(want) != len(got) {
		Ko(t, "Expected %d registrations but got %d", len(want), len(got))
	}

outer:
	for _, rw := range want {
		for _, rg := range got {
			if rg.DevEUI() != rw.DevEUI {
				Ko(t, "Expected registration for %v but got for %v", rw.DevEUI, rg.DevEUI())
			}
			if reflect.DeepEqual(rw.Recipient.Recipient, rg.Recipient()) {
				continue outer
			}
		}
		Ko(t, "Registrations don't match expectation.\nWant: %v\nGot:  %v", want, got)
	}
	Ok(t, "Check registrations")
}

func checkPayloads(t *testing.T, want string, got []string) {
	for _, payload := range got {
		if want != payload {
			Ko(t, "Paylaod don't match expectation.\nWant: %s\nGot:  %s", want, payload)
		}
	}
	Ok(t, "Check payloads")
}

func CheckResps(t *testing.T, want *MsgRes, got chan MsgRes) {
	if want == nil {
		if len(got) == 0 {
			Ok(t, "Check Resps")
			return
		}
		Ko(t, "Expected no message response but got one")
	}

	if len(got) < 1 {
		Ko(t, "Expected one message but got none")
	}

	msg := <-got
	mocks.Check(t, *want, msg, "Resps")
}

func CheckRecipients(t *testing.T, want Recipient, got Recipient) {
	mocks.Check(t, want, got, "Recipients")
}

func CheckJSONs(t *testing.T, want []byte, got []byte) {
	mocks.Check(t, want, got, "JSON")
}

func CheckMethods(t *testing.T, want string, got string) {
	mocks.Check(t, want, got, "Methods")
}

func CheckContentTypes(t *testing.T, want string, got string) {
	mocks.Check(t, want, got, "ContentTypes")
}
