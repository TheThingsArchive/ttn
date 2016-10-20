// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/viper"
)

func PrintConfig(ctx log.Interface, debug bool) {
	prt := ctx.Infof
	if debug {
		prt = ctx.Debugf
	}

	prt("Using config:")
	fmt.Println()
	printKV("config file", viper.ConfigFileUsed())
	printKV("data dir", viper.GetString("data"))
	fmt.Println()

	for key, val := range viper.AllSettings() {
		switch key {
		case "builddate":
			fallthrough
		case "gitcommit":
			fallthrough
		case "gitbranch":
			fallthrough
		case "version":
			continue
		default:
			printKV(key, val)
		}
	}
	fmt.Println()
}

func printKV(key, val interface{}) {
	fmt.Printf("%20s: %s\n", key, val)
}
