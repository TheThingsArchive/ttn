// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"testing"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	tt "github.com/TheThingsNetwork/ttn/utils/testing"
)

func getLogger(t *testing.T, tag string) ttnlog.Interface {
	return tt.GetLogger(t, tag)
}
