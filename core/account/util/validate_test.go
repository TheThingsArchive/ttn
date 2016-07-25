// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

type Foo struct {
	Foo string `valid:"required"`
}

var (
	validFoo = Foo{
		Foo: "Hi",
	}

	invalidFoo = Foo{}

	validFoos = []Foo{
		validFoo,
		validFoo,
	}

	invalidFoos = []Foo{
		validFoo,
		invalidFoo,
	}
)

func TestValidateStruct(t *testing.T) {
	a := New(t)

	err := Validate(validFoo)
	a.So(err, ShouldBeNil)

	err = Validate(invalidFoo)
	a.So(err, ShouldNotBeNil)
}

func TestValidateSlice(t *testing.T) {
	a := New(t)

	err := Validate(validFoos)
	a.So(err, ShouldBeNil)

	err = Validate(invalidFoos)
	a.So(err, ShouldNotBeNil)
}

func TestValidatePtr(t *testing.T) {
	a := New(t)

	err := Validate(&validFoo)
	a.So(err, ShouldBeNil)

	err = Validate(&invalidFoo)
	a.So(err, ShouldNotBeNil)
}

func TestValidatePtrSlice(t *testing.T) {
	a := New(t)

	err := Validate(&validFoos)
	a.So(err, ShouldBeNil)

	err = Validate(&invalidFoos)
	a.So(err, ShouldNotBeNil)
}
