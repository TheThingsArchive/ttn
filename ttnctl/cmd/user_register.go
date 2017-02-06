// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
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
		assertArgsLength(cmd, args, 2, 2)

		username := args[0]
		email := args[1]
		fmt.Print("Password: ")
		password, err := gopass.GetPasswd()
		if err != nil {
			ctx.Fatal(err.Error())
		}
		acc := account.New(viper.GetString("auth-server")).WithHeader("User-Agent", util.GetUserAgent())
		err = acc.RegisterUser(username, email, string(password))
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
