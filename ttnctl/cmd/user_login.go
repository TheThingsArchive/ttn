// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
)

var userLoginCmd = &cobra.Command{
	Use:   "login [e-mail]",
	Short: "Login",
	Long:  `ttnctl user login allows you to login to the account server`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.UsageFunc()(cmd)
			return
		}

		email := args[0]
		fmt.Print("Password: ")
		password, err := gopass.GetPasswd()
		if err != nil {
			ctx.Fatal(err.Error())
		}

		util.Login(ctx, email, string(password))
	},
}

func init() {
	userCmd.AddCommand(userLoginCmd)
}
