// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// usersCmd represents the users command
var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users on the account server",
	Long:  `ttnctl users allows you to manage users`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var usersCreateCmd = &cobra.Command{
	Use:   "create [e-mail]",
	Short: "Create a new user",
	Long:  `ttnctl users create allows you to create a new user`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			ctx.Fatal("Insufficient arguments")
		}

		email := args[0]
		fmt.Print("Password: ")
		password, err := gopass.GetPasswd()
		if err != nil {
			ctx.Fatal("Invalid password")
		}

		uri := fmt.Sprintf("http://%s/register", viper.GetString("ttn-account-server"))
		values := url.Values{
			"email":    {email},
			"password": {string(password)},
		}
		res, err := http.PostForm(uri, values)
		if err != nil {
			ctx.WithError(err).Fatal("Registration failed")
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			ctx.Fatalf("Registration failed: %d %s", res.StatusCode, res.Status)
		}

		ctx.Info("User created")
	},
}

func init() {
	RootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(usersCreateCmd)

	usersCmd.PersistentFlags().String("ttn-account-server", "account.thethings.network", "The Things Network OAuth 2.0 account server")
	viper.BindPFlag("ttn-account-server", usersCmd.PersistentFlags().Lookup("ttn-account-server"))
}
