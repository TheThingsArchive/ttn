// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"github.com/thethingsnetwork/core"
	. "github.com/thethingsnetwork/core/utils/testing"
	"testing"
)

func TestNewAdapter(t *testing.T) {
	Ok(t, "pending")
}

func TestSend(t *testing.T) {
	Desc(t, "Send is not supported")
	adapter, err := NewAdapter(33000)
	if err != nil {
		panic(err)
	}
	err = adapter.Send(core.Packet{})
	checkErrors(t, ErrNotSupported, err)
}

func TestNextRegistration(t *testing.T) {
	Desc(t, "Next registration is not supported")
	adapter, err := NewAdapter(33001)
	if err != nil {
		panic(err)
	}
	_, _, err = adapter.NextRegistration()
	checkErrors(t, ErrNotSupported, err)
}

func TestNext(t *testing.T) {
	Ok(t, "pending")
}

func checkErrors(t *testing.T, want error, got error) {
	if want == got {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be %v but got %v", want, got)
}
