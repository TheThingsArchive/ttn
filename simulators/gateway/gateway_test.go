// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNew(t *testing.T) {
	Convey("The New method should return a valid gateway struct ready to use", t, func() {
		id := "qwerty"
		router1 := "router1Addr"
		router2 := "router2Addr"

		Convey("Given an identifier and a router address", func() {
			gateway, err := New(id, router1)

			Convey("No error should have been trown", func() {
				So(err, ShouldBeNil)
			})

			Convey("The identifier should have been set correctly", func() {
				So(gateway.Id, ShouldEqual, id)
			})

			Convey("The list of configured routers should have been set correctly", func() {
				So(gateway.routers, ShouldResemble, []string{router1})
			})
		})

		Convey("Given an identifier and several routers address", func() {
			gateway, err := New(id, router1, router2)

			Convey("No error should have been trown", func() {
				So(err, ShouldBeNil)
			})

			Convey("The identifier should have been set correctly", func() {
				So(gateway.Id, ShouldEqual, id)
			})

			Convey("The list of configured routers should have been set correctly", func() {
				So(gateway.routers, ShouldResemble, []string{router1, router2})
			})
		})

		Convey("Given a bad identifier and/or bad router addresses", func() {
			Convey("It should return an error for an empty id", func() {
				gateway, err := New("", router1)
				So(gateway, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("It should return an error for an empty routers list", func() {
				gateway, err := New(id)
				So(gateway, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("It should return an error for an invalid router address", func() {
				gateway, err := New(id, "")
				So(gateway, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})
}
