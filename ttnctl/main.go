// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/cmd"
	"github.com/spf13/viper"
)

var (
	version   = "2.x.x"
	gitBranch = "unknown"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	viper.Set("version", version)
	viper.Set("gitBranch", gitBranch)
	viper.Set("gitCommit", gitCommit)
	viper.Set("buildDate", buildDate)
	cmd.Execute()
}
