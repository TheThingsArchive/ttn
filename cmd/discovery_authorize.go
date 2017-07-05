// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/ttn/utils/security"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var discoveryAuthorizeCmd = &cobra.Command{
	Hidden: true,
	Use:    "authorize [router/broker/handler] [id]",
	Short:  "Generate a token that components should use to announce themselves",
	Long:   `ttn discovery authorize generates a token that components should use to announce themselves`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.UsageFunc()(cmd)
			return
		}

		privKey, err := security.LoadKeypair(viper.GetString("key-dir"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not load security keys")
		}

		ttl, err := cmd.Flags().GetInt("valid")
		if err != nil {
			ctx.WithError(err).Fatal("Could not read TTL")
		}

		issuer, err := cmd.Flags().GetString("issuer")
		if err != nil {
			ctx.WithError(err).Fatal("Could not read issuer ID")
		}

		var claims claims.ComponentClaims
		claims.Subject = args[1]
		claims.Type = args[0]
		claims.Issuer = issuer
		claims.IssuedAt = time.Now().Unix()
		claims.NotBefore = time.Now().Unix()
		if ttl > 0 {
			claims.ExpiresAt = time.Now().Add(time.Duration(ttl) * time.Hour * 24).Unix()
		}
		tokenBuilder := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		token, err := tokenBuilder.SignedString(privKey)
		if err != nil {
			ctx.WithError(err).Fatal("Could not sign JWT")
		}

		ctx.WithField("ID", args[0]).Info("Generated token")

		if filepath := viper.GetString("save"); filepath != "" {
			ctx := ctx.WithField("Filepath", filepath)
			f, err := os.Create(filepath)
			if err != nil {
				ctx.WithError(err).Error("Could not save token in specified file")
			} else {
				defer f.Close()
				if _, err = f.Write([]byte(token)); err != nil {
					ctx.WithError(err).Error("Could not write token in specified file")
				} else {
					ctx.Info("Token saved in specified file")
				}
			}
		}

		fmt.Println()
		fmt.Println(token)
		fmt.Println()
	},
}

func init() {
	discoveryCmd.AddCommand(discoveryAuthorizeCmd)
	discoveryAuthorizeCmd.Flags().Int("valid", 0, "The number of days the token is valid")
	discoveryAuthorizeCmd.Flags().String("issuer", "local", "The issuer ID to use")
	discoveryAuthorizeCmd.Flags().String("save", "", "If you wish to store the token in a file, path to the file where the token will be saved")
}
