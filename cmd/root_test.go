package cmd

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestRootCmd(t *testing.T) {
	a := New(t)
	a.So(RootCmd.IsAvailableCommand(), ShouldBeTrue)
}
