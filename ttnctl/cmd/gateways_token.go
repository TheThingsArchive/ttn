// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var gatewaysTokenCmd = &cobra.Command{
	Use:    "token [Type] [gatewayID]",
	Hidden: true,
	Short:  "Get the token for a gateway.",
	Long:   `gateways token gets a signed token for the gateway.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return
		}

		gatewayID := args[0]
		if !api.ValidID(gatewayID) {
			ctx.Fatal("Invalid Gateway ID")
		}
		ctx = ctx.WithField("id", gatewayID)

		account := util.GetAccount(ctx)

		token, err := account.GetGatewayToken(gatewayID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get gateway token")
		}
		if token.Token == "" {
			ctx.Fatal("Gateway token was empty")
		}

		ctx.Info("Got gateway token")

		fmt.Println()
		fmt.Println(token.Token)
		fmt.Println()
		if !token.Expires.IsZero() {
			fmt.Printf("Expires %s\n", token.Expires)
			fmt.Println()
		}
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysTokenCmd)
}
