// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestRootCmd(t *testing.T) {
	a := New(t)
	a.So(RootCmd.IsAvailableCommand(), ShouldBeTrue)
}
