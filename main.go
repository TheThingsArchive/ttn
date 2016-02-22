// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"github.com/TheThingsNetwork/ttn/cmd"
	"github.com/spf13/viper"
)

var (
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	viper.Set("version", "a development version")
	viper.Set("git-commit", gitCommit)
	viper.Set("build-date", buildDate)
	cmd.Execute()
}
