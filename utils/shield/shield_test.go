// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package shield

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestShield(t *testing.T) {
	{
		shield := New(2)
		err1 := shield.ThroughIn()
		err2 := shield.ThroughIn()
		err3 := shield.ThroughIn()
		shield.ThroughOut()
		err4 := shield.ThroughIn()
		CheckErrors(t, nil, err1)
		CheckErrors(t, nil, err2)
		CheckErrors(t, ErrOperational, err3)
		CheckErrors(t, nil, err4)
	}

	// ----------

	{
		shield := New(1)
		shield.ThroughOut()
		err1 := shield.ThroughIn()
		err2 := shield.ThroughIn()
		CheckErrors(t, nil, err1)
		CheckErrors(t, ErrOperational, err2)
	}
}
