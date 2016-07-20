// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	s "github.com/smartystreets/assertions"
)

func TestCollaboratorRights(t *testing.T) {
	a := s.New(t)
	c := Collaborator{
		Username: "username",
		Rights: []types.Right{
			types.Right("right"),
		},
	}

	a.So(c.HasRight(types.Right("right")), s.ShouldBeTrue)
	a.So(c.HasRight(types.Right("foo")), s.ShouldBeFalse)
}
