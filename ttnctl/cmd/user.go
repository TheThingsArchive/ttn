// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Show the current user",
	Long:  `ttnctl user shows the current logged on user's profile`,
	Example: `$ ttnctl user
  INFO Found user profile:

            Username: yourname
                Name: Your Name
               Email: your@email.org

  INFO Login credentials valid until Sep 20 09:04:12
`,
	Run: func(cmd *cobra.Command, args []string) {
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
		tokenParts := strings.Split(token.AccessToken, ".")
		if len(tokenParts) != 3 {
			ctx.Fatal("Invalid access token")
		}
		segment, err := jwt.DecodeSegment(tokenParts[1])
		if err != nil {
			ctx.WithError(err).Fatal("Could not decode access token")
		}
		var claims claims.Claims
		err = json.Unmarshal(segment, &claims)
		if err != nil {
			ctx.WithError(err).Fatal("Could not unmarshal access token")
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
