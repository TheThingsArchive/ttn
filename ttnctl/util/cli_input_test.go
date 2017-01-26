package util

import (
	"errors"
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestIsErrOutOfRange(t *testing.T) {
	a := New(t)

	err := errors.New("Dummy error")
	ok := IsErrOutOfRange(err)
	a.So(ok, ShouldBeFalse)

	err = errors.New("value out of range")
	ok = IsErrOutOfRange(err)
	a.So(ok, ShouldBeTrue)

	err = errors.New("whatever message: value out of range")
	ok = IsErrOutOfRange(err)
	a.So(ok, ShouldBeTrue)
}
