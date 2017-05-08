// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"errors"
	"testing"

	"github.com/smartystreets/assertions"
)

func resetInitFuncs() {
	_initFuncs = make(map[int][]initFunc)
}

func TestInit(t *testing.T) {
	a := assertions.New(t)

	resetInitFuncs()
	c := new(Component)

	var called int
	OnInitialize(func(c *Component) error {
		called++
		return nil
	})

	a.So(c.initialize(), assertions.ShouldBeNil)
	a.So(called, assertions.ShouldEqual, 1)

	resetInitFuncs()
	c = new(Component)

	OnInitialize(func(c *Component) error { return errors.New("err") })
	a.So(c.initialize(), assertions.ShouldNotBeNil)
}
