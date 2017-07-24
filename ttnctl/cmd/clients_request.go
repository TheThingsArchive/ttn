// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	accountlib "github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var clientRequestCmd = &cobra.Command{
	Use:   "request [Name] [Description]",
	Short: "Request a client",
	Long:  "ttnctl clients request can be used to request an OAuth client from the network staff.",
	Example: `$ ttnctl clients request my-gateway-editor "Client used to consult and edit gateway information" --uri "https://mygatewayclient.org/oauth/callback" --scope "profile,gateways" --grants "authorization_code,refresh_token"
  INFO OAuth client requested OAuthClientName=my-gateway-editor
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)

		var name = args[0]
		var description = args[1]

		ctx = ctx.WithField("OAuthClientName", name)

		uri, err := cmd.Flags().GetString("uri")
		if err != nil {
			ctx.WithError(err).Fatal("Error with URI")
		}

		scopes := make([]string, 0)
		strScopes, err := cmd.Flags().GetStringSlice("scope")
		if err != nil {
			ctx.WithError(err).Fatal("Could not parse scope")
		}
		for _, strScope := range strScopes {
			scopes = append(scopes, strScope)
		}

		strGrants, err := cmd.Flags().GetStringSlice("grants")
		if err != nil {
			ctx.WithError(err).Fatal("Could not parse grants")
		}
		grants := make([]accountlib.Grant, 0)
		for _, strGrant := range strGrants {
			grants = append(grants, accountlib.Grant(strGrant))
		}

		account := util.GetAccount(ctx)

		request := accountlib.OAuthClient{
			Name:        name,
			Description: description,
			URI:         uri,
			Grants:      grants,
			Scope:       scopes,
		}

		_, err = account.CreateOAuthClient(&request)
		if err != nil {
			ctx.WithError(err).Fatal("Could not request OAuth Client")
		}

		util.ForceRefreshToken(ctx)

		ctx.Info("OAuth client requested")
	},
}

func init() {
	clientsCmd.AddCommand(clientRequestCmd)
	clientRequestCmd.Flags().String("uri", "", "Pass a callback URI for the OAuth client")
	clientRequestCmd.Flags().StringSlice("scope", []string{}, "Scopes requested for the OAuth client")
	clientRequestCmd.Flags().StringSlice("grants", []string{}, "Grants requested for the OAuth client")
}
