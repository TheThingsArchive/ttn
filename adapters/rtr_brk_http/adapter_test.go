// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package rtr_brk_http

import (
	//	"fmt"
	"github.com/thethingsnetwork/core"
	//	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/testing/mock_components"
	"github.com/thethingsnetwork/core/utils/log"
	. "github.com/thethingsnetwork/core/utils/testing"
	//	"net/http"
	//	"reflect"
	"testing"
	//	"time"
)

// ----- The adapter can be created and listen straigthforwardly
func TestListenOptionsTest(t *testing.T) {
	adapter, router := generateAdapterAndRouter(t)

	Desc(t, "Listen to adapter")
	if err := adapter.Listen(router, nil); err != nil {
		Ko(t, "No error was expected but got: %+v", err)
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
