// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var userLoginCmd = &cobra.Command{
	Use:   "login [client code]",
	Short: "Login",
	Long:  `ttnctl user login allows you to login to the account server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.UsageFunc()(cmd)
			return
		}

		code := args[0]
		util.Login(ctx, code)
	},
}

func init() {
	userCmd.AddCommand(userLoginCmd)
}
