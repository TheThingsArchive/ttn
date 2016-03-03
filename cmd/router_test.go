package cmd

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestRouterCmd(t *testing.T) {
	a := New(t)
	a.So(routerCmd.IsAvailableCommand(), ShouldBeTrue)
}
