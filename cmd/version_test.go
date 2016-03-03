package cmd

import (
	"testing"

	"github.com/apex/log"
	"github.com/apex/log/handlers/memory"
	. "github.com/smartystreets/assertions"
	"github.com/spf13/viper"
)

func TestVersionCmd(t *testing.T) {
	a := New(t)

	h := memory.New()

	ctx = &log.Logger{
		Level:   log.DebugLevel,
		Handler: h,
	}

	viper.Set("version", "v-test")
	versionCmd.Run(versionCmd, []string{})

	a.So(h.Entries, ShouldHaveLength, 1)
	a.So(h.Entries[0].Message, ShouldContainSubstring, "The Things Network")
	a.So(h.Entries[0].Message, ShouldContainSubstring, "v-test")
}
