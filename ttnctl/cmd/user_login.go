// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/account"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var userLoginCmd = &cobra.Command{
	Use:   "login [client code]",
	Short: "Login",
	Long:  `ttnctl user login allows you to login to the account server.`,
	Example: `First get an access code from your TTN Profile by going to
https://account.thethingsnetwork.org and clicking "ttnctl access code".

$ ttnctl user login 2keK3FTu6e0327cq4ni0wRTMT2mTS-m_FLzFBlNQadwa
  INFO Successfully logged in as yourname (your@email.org)
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.UsageFunc()(cmd)
			return
		}

		code := args[0]
		token, err := util.Login(ctx, code)
		if err != nil {
			ctx.WithError(err).Fatal("Login failed")
		}

		acc := account.New(viper.GetString("ttn-account-server"), token.AccessToken)
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
