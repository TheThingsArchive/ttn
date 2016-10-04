// +build !homebrew

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/utils/version"
	"github.com/spf13/cobra"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "selfupdate",
	Short: "Update ttnctl to the latest version",
	Long:  `ttnctl selfupdate updates the current ttnctl to the latest version`,
	Run: func(cmd *cobra.Command, args []string) {
		version.Selfupdate(ctx, "ttnctl")
	},
}

func init() {
	RootCmd.AddCommand(selfUpdateCmd)
}
