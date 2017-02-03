// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"users"},
	Short:   "Show the current user",
	Long:    `ttnctl user shows the current logged on user's profile`,
	Example: `$ ttnctl user
  INFO Found user profile:

            Username: yourname
                Name: Your Name
               Email: your@email.org

  INFO Login credentials valid until Sep 20 09:04:12
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 0, 0)

		account := util.GetAccount(ctx)
		profile, err := account.Profile()
		if err != nil {
			ctx.WithError(err).Fatal("Could not get user profile")
		}

		ctx.Info("Found user profile:")
		fmt.Println()
		printKV("Username", profile.Username)
		printKV("Name", profile.Name)
		printKV("Email", profile.Email)
		fmt.Println()

		token, err := util.GetTokenSource(ctx).Token()
		if err != nil {
			ctx.WithError(err).Fatal("Could not get access token")
		}

		claims, err := claims.FromTokenWithoutValidation(token.AccessToken)
		if err != nil {
			ctx.WithError(err).Fatal("Could not parse token")
		}

		if claims.ExpiresAt != 0 {
			expires := time.Unix(claims.ExpiresAt, 0)
			if expires.After(time.Now()) {
				ctx.Infof("Login credentials valid until %s", expires.Format(time.Stamp))
			} else {
				ctx.Warnf("Login credentials expired %s", expires.Format(time.Stamp))
			}
		}

	},
}

func init() {
	RootCmd.AddCommand(userCmd)
}
