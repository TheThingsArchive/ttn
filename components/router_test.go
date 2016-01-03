// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/thethingsnetwork/core"
	. "github.com/thethingsnetwork/core/utils/testing"
	"testing"
)

// ----- A new router instance can be created an obtained from a constuctor
func TestNewRouter(t *testing.T) {
	tests := []newRouterTest{
		{genBrokers(), nil},
		{[]core.BrokerAddress{}, core.ErrBadOptions},
	}

	for _, test := range tests {
		test.run(t)
	}
}

type newRouterTest struct {
	in   []core.BrokerAddress
	want error
}

func (test newRouterTest) run(t *testing.T) {
	Desc(t, "Create new router with params: %v", test.in)
	router, err := NewRouter(test.in...)
	checkErrors(t, test.want, err, router)
}

// ----- Build Utilities
func genBrokers() []core.BrokerAddress {
	return []core.BrokerAddress{
		core.BrokerAddress("0.0.0.0:3000"),
		core.BrokerAddress("0.0.0.0:3001"),
	}
}

// ----- Check Utilities
func checkErrors(t *testing.T, want error, got error, router core.Router) {
	if want != got {
		Ko(t, "Expected error {%v} but got {%v}", want, got)
		return
	}

	if want == nil && router == nil {
		Ko(t, "Expected no error but got a nil router")
		return
	}

	Ok(t)
}
