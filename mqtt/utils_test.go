// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"testing"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	tt "github.com/TheThingsNetwork/ttn/utils/testing"
)

func getLogger(t *testing.T, tag string) log.Interface {
	return apex.Wrap(tt.GetLogger(t, tag))
}
