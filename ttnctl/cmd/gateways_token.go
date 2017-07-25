// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/api"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var gatewaysTokenCmd = &cobra.Command{
	Use:    "token [GatewayID]",
	Hidden: true,
	Short:  "Get the token for a gateway.",
	Long:   `gateways token gets a signed token for the gateway.`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		gatewayID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(gatewayID, "Gateway ID"); err != nil {
			ctx.Fatal(err.Error())
		}
		ctx = ctx.WithField("id", gatewayID)

		account := util.GetAccount(ctx)

		token, err := account.GetGatewayToken(gatewayID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get gateway token")
		}
		if token.AccessToken == "" {
			ctx.Fatal("Gateway token was empty")
		}

		ctx.Info("Got gateway token")

		fmt.Println()
		fmt.Println(token.AccessToken)
		fmt.Println()
		if !token.Expiry.IsZero() {
			fmt.Printf("Expires %s\n", token.Expiry)
			fmt.Println()
		}
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysTokenCmd)
}
