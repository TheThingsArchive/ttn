// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// userCmd represents the users command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Show the current user",
	Long:  `ttnctl user shows the current logged on user`,
	Run: func(cmd *cobra.Command, args []string) {
		t, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication token")
		}
		if t == nil {
			ctx.Warn("No login found. Please login with ttnctl user login [e-mail]")
			return
		}

		ctx.Infof("Logged on as %s", t.Email)
	},
}

var userCreateCmd = &cobra.Command{
	Use:   "create [e-mail]",
	Short: "Create a new user",
	Long:  `ttnctl user create allows you to create a new user`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		email := args[0]
		fmt.Print("Password: ")
		password, err := gopass.GetPasswd()
		if err != nil {
			ctx.Fatal(err.Error())
		}
		fmt.Print("Confirm password: ")
		password2, err := gopass.GetPasswd()
		if err != nil {
			ctx.Fatal(err.Error())
		}
		if !bytes.Equal(password, password2) {
			ctx.Fatal("Passwords do not match")
		}

		uri := fmt.Sprintf("%s/register", viper.GetString("ttn-account-server"))
		values := url.Values{
			"email":    {email},
			"password": {string(password)},
		}
		res, err := http.PostForm(uri, values)
		if err != nil {
			ctx.WithError(err).Fatal("Registration failed")
		}

		if res.StatusCode != http.StatusCreated {
			ctx.Fatalf("Registration failed: %d %s", res.StatusCode, res.Status)
		}

		ctx.Info("User created")
	},
}

var userLoginCmd = &cobra.Command{
	Use:   "login [e-mail]",
	Short: "Login",
	Long:  `ttnctl user login allows you to login`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		email := args[0]
		fmt.Print("Password: ")
		password, err := gopass.GetPasswd()
		if err != nil {
			ctx.Fatal(err.Error())
		}

		_, err = util.Login(viper.GetString("ttn-account-server"), email, string(password))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to login")
		}

		ctx.Infof("Logged in as %s and persisted token in %s", email, util.AuthsFileName)
	},
}

var userLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout the current user",
	Long:  `ttnctl user logout logs out the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := util.Logout(viper.GetString("ttn-account-server")); err != nil {
			ctx.WithError(err).Fatal("Failed to log out")
		}

		ctx.Info("Logged out")
	},
}

func init() {
	RootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userLoginCmd)
	userCmd.AddCommand(userLogoutCmd)
}
