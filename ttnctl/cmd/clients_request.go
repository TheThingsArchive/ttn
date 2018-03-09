// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"net/http"

	accountlib "github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var clientRequestCmd = &cobra.Command{
	Use:   "request [Name] [Description]",
	Short: "Request a client",
	Long: `ttnctl clients request can be used to request an OAuth client from the network staff.
	You need to supply the following information:
	- An identifier for your OAuth client; can contain lowercase letters, numbers, dashes and underscores, just like Application and Gateway IDs.
	- A description that will be shown to users that are signing in.
	- A callback URI where users will be redirected after login.
	- The scopes that your client needs access to:
		- apps: Create and delete Applications
		- gateways: Create and delete Gateways
		- profile: Edit user profiles
		- Note that you may not need an OAuth client to manage devices.
	- The grants that your client uses for login:
		- authorization_code: OAuth 2.0 authorization code (this is probably what you need)
		- refresh_token: OAuth 2.0 refresh token grant
		- password: OAuth 2.0 password grant (this will usually not be accepted)`,
	Example: `$ ttnctl clients request my-gateway-editor "Client used to consult and edit gateway information" --uri "https://mygatewayclient.org/oauth/callback" --scope "profile,gateways" --grants "authorization_code,refresh_token"
  INFO OAuth client requested OAuthClientName=my-gateway-editor
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)

		var name = args[0]
		var description = args[1]

		uri, err := cmd.Flags().GetString("uri")
		if err != nil {
			ctx.WithError(err).Fatal("Error with URI")
		}

		var uriOK bool
		ctx.Info("Testing Callback URI: " + uri + "?code=test&state=test")
		res, err := http.Get(uri + "?code=test&state=test")
		switch {
		case err != nil:
			ctx.WithError(err).Error("Callback URI test failed.")
		case res.StatusCode == 404:
			ctx.Error("Callback URI was not found (404)")
		case res.StatusCode >= 500:
			ctx.Errorf("Callback URI errored (%d)", res.StatusCode)
		default:
			ctx.Infof("Callback URI seems to be reachable (returned %d)", res.StatusCode)
			uriOK = true
		}
		if !uriOK && !confirm("Are you sure the URI is correct? (y/N)") {
			ctx.Info("Aborting")
			return
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
