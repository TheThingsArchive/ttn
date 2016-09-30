// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "Manage applications",
	Long:  `ttnctl applications can be used to manage applications.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)
		util.GetAccount(ctx)
	},
}

func init() {
	RootCmd.AddCommand(applicationsCmd)
}
