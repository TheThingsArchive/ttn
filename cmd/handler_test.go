package cmd

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestHandlerCmd(t *testing.T) {
	a := New(t)
	a.So(handlerCmd.IsAvailableCommand(), ShouldBeTrue)
}
