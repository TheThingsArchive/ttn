// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/go-account-lib/auth"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var userLoginCmd = &cobra.Command{
	Use:   "login [access code]",
	Short: "Log in with your TTN account",
	Long:  `ttnctl user login allows you to log in to your TTN account.`,
	Example: `First get an access code from your TTN profile by going to
https://account.thethingsnetwork.org and clicking "ttnctl access code".

$ ttnctl user login [paste the access code you requested above]
  INFO Successfully logged in as yourname (your@email.org)
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		code := args[0]
		token, err := util.Login(ctx, code)
		if err != nil {
			ctx.WithError(err).Fatal("Login failed")
		}

		acc := account.New(viper.GetString("auth-server")).WithHeader("User-Agent", util.GetUserAgent())
		acc.WithAuth(auth.AccessToken(token.AccessToken))
		profile, err := acc.Profile()
		if err != nil {
			ctx.WithError(err).Fatal("Could not get user profile")
		}
		ctx.Info(fmt.Sprintf("Successfully logged in as %s (%s)", profile.Username, profile.Email))
	},
}

func init() {
	userCmd.AddCommand(userLoginCmd)
}
