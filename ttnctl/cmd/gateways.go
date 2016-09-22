// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import "github.com/spf13/cobra"

var gatewaysCmd = &cobra.Command{
	Use:   "gateways",
	Short: "Manage gateways",
	Long:  `ttnctl gateways can be used to manage gateways.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)
	},
}

func init() {
	RootCmd.AddCommand(gatewaysCmd)
}
