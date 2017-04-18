// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
  "github.com/TheThingsNetwork/ttn/api/router"
  "github.com/spf13/viper"
)

func main() {
	viper.SetDefault("ttn-account-server", "https://account.thethingsnetwork.org")
	viper.SetDefault("ttn-handler",        "staging.thethingsnetwork.org:1782")
	viper.SetDefault("mqtt-broker",        "staging.thethingsnetwork.org:1883")
	router.Start()
}
