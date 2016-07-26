// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/security"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// networkserverAuthorizeCmd represents the secure command
var networkserverAuthorizeCmd = &cobra.Command{
	Use:   "authorize [id]",
	Short: "Generate a token that Brokers should use to connect",
	Long:  `ttn networkserver authorize generates a token that Brokers should use to connect`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
		}

		_, priv, _, err := security.LoadKeys(viper.GetString("key-dir"))
		if err != nil {
			ctx.WithError(err).Fatal("Could not load security keys")
		}
		ttl, err := cmd.Flags().GetInt("valid")
		if err != nil {
			ctx.WithError(err).Fatal("Could not read TTL")
		}
		token, err := security.BuildJWT(args[0], time.Duration(ttl)*time.Hour*24, priv)
		if err != nil {
			ctx.WithError(err).Fatal("Could not generate a JWT")
		}

		ctx.WithField("ID", args[0]).Info("Generated NS token")
		fmt.Println()
		fmt.Println(token)
		fmt.Println()
	},
}

func init() {
	networkserverCmd.AddCommand(networkserverAuthorizeCmd)
	networkserverAuthorizeCmd.Flags().Int("valid", 0, "The number of days the token is valid")
}
