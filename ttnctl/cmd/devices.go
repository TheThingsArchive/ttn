// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import "github.com/spf13/cobra"

// devicesCmd is the entrypoint for handlerctl
var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage devices",
	Long:  `ttnctl devices can be used to manage devices.`,
}

func init() {
	RootCmd.AddCommand(devicesCmd)
}
