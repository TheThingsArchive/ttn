package cmd

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestBrokerCmd(t *testing.T) {
	a := New(t)
	a.So(brokerCmd.IsAvailableCommand(), ShouldBeTrue)
}
