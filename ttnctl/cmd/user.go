// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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

		req, err := http.NewRequest("POST", uri, strings.NewReader(values.Encode()))
		if err != nil {
			ctx.WithError(err).Fatalf("Failed to create request")
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			ctx.WithError(err).Fatal("Registration failed")
		}

		if res.StatusCode != http.StatusCreated {
			buf, _ := ioutil.ReadAll(res.Body)
			ctx.Fatalf("Registration failed: %s (%v)", res.Status, string(buf))
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

		server := viper.GetString("ttn-account-server")
		_, err = util.Login(server, email, string(password))
		if err != nil {
			ctx.Info(fmt.Sprintf("Visit %s to register or to retrieve your account credentials.", server))
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
