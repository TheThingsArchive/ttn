// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/account"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var userRegisterCmd = &cobra.Command{
	Use:   "register [username] [e-mail]",
	Short: "Register",
	Long:  `ttnctl user register allows you to register a new user in the account server`,
	Example: `$ ttnctl user register yourname your@email.org
Password: <entering password>
  INFO Registered user
  WARN You might have to verify your email before you can login
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.UsageFunc()(cmd)
			return
		}
		username := args[0]
		email := args[1]
		fmt.Print("Password: ")
		password, err := gopass.GetPasswd()
		if err != nil {
			ctx.Fatal(err.Error())
		}
		err = account.RegisterUser(viper.GetString("ttn-account-server"), username, email, string(password))
		if err != nil {
			ctx.WithError(err).Fatal("Could not register user")
		}
		ctx.Info("Registered user")
		ctx.Warn("You might have to verify your email before you can login")
	},
}

func init() {
	userCmd.AddCommand(userRegisterCmd)
}
