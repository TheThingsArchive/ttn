// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type token struct {
	AccessToken      string `json:"access_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// usersCmd represents the users command
var usersCmd = &cobra.Command{
	Use:   "user",
	Short: "Show the current user",
	Long:  `ttnctl user shows the current logged on user`,
	Run: func(cmd *cobra.Command, args []string) {
		t, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}

		if t == nil {
			ctx.Warn("No login found")
			return
		}

		// TODO: Validate token

		ctx.Infof("Logged on as %s", t.Email)
	},
}

var usersCreateCmd = &cobra.Command{
	Use:   "create [e-mail]",
	Short: "Create a new user",
	Long:  `ttnctl user create allows you to create a new user`,
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

		uri := fmt.Sprintf("%s/register", viper.GetString("ttn-account-server"))
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

var usersLoginCmd = &cobra.Command{
	Use:   "login [e-mail]",
	Short: "Login",
	Long:  `ttnctl user login allows you to login`,
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

		server := viper.GetString("ttn-account-server")
		uri := fmt.Sprintf("%s/token", server)
		values := url.Values{
			"grant_type": {"password"},
			"username":   {email},
			"password":   {string(password)},
		}
		req, err := http.NewRequest("POST", uri, strings.NewReader(values.Encode()))
		if err != nil {
			ctx.WithError(err).Fatal("Create request failed")
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("ttnctl", "")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			ctx.WithError(err).Fatal("Request failed")
		}
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		var t token
		if err := decoder.Decode(&t); err != nil {
			ctx.WithError(err).Fatal("Failed to parse response")
		}

		if resp.StatusCode != http.StatusOK {
			if t.Error != "" {
				ctx.Fatalf("Request failed: %s", t.ErrorDescription)
			} else {
				ctx.Fatalf("Request failed: %s", resp.Status)
			}
		}

		ctx.Infof("Logged in as %s", email)

		if err := util.SaveAuth(server, email, t.AccessToken); err != nil {
			ctx.WithError(err).Error("Failed to save login")
		}
	},
}

func init() {
	RootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(usersCreateCmd)
	usersCmd.AddCommand(usersLoginCmd)

	usersCmd.PersistentFlags().String("ttn-account-server", "https://account.thethings.network", "The Things Network OAuth 2.0 account server")
	viper.BindPFlag("ttn-account-server", usersCmd.PersistentFlags().Lookup("ttn-account-server"))
}
