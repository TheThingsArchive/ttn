package http

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/smartystreets/assertions"
)

func TestBadRequest(t *testing.T) {
	rw := ResponseWriter{}
	BadRequest(&rw, "Test")
	a := assertions.New(t)
	a.So(rw.TheStatus, assertions.ShouldEqual, 400)
	a.So(string(rw.TheBody), assertions.ShouldEqual, "Test")
}
