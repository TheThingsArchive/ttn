// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gtw_rtr_udp

import (
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/testing/mock_components"
	"github.com/thethingsnetwork/core/utils/log"
	. "github.com/thethingsnetwork/core/utils/testing"
	"testing"
)

// ----- The adapter should be able to create a udp connection given a valid udp port
func TestListenOptions(t *testing.T) {
	tests := []listenOptionsTest{
		{uint(3000), nil},
		{uint(3000), core.ErrBadOptions}, // Already used now
		{int(14), core.ErrBadOptions},
		{"somethingElse", core.ErrBadOptions},
	}

	for _, test := range tests {
		test.run(t)
	}
}

type listenOptionsTest struct {
	options interface{}
	want    error
}

func (test listenOptionsTest) run(t *testing.T) {
	Desc(t, "Run Listen(router, %T %v)", test.options, test.options)
	adapter, router := generateAdapterAndRouter(t)
	got := adapter.Listen(router, test.options)
	test.check(t, got) // Check if errors match
}

func (test listenOptionsTest) check(t *testing.T, got error) {
	if got != test.want {
		t.Errorf("expected {%v} to be {%v}\n", got, test.want)
		Ko(t)
		return
	}
	Ok(t)
}

// ----- Build Utilities
func generateAdapterAndRouter(t *testing.T) (Adapter, core.Router) {
	return Adapter{
		Logger: log.TestLogger{
			Tag: "Adapter",
			T:   t,
		},
	}, mock_components.NewRouter()
}

// ----- Operate Utilities

// ----- Check Utilities
func checkListenResult(t *testing.T, got error, wanted error, options interface{}) {
}
