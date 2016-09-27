// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/cmd"
	"github.com/spf13/viper"
)

var (
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	viper.Set("version", "2.0.0-dev")
	viper.Set("gitCommit", gitCommit)
	viper.Set("buildDate", buildDate)
	cmd.Execute()
}
