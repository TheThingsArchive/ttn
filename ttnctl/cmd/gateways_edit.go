// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"strings"

	"github.com/TheThingsNetwork/api"
	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/spf13/cobra"
)

var gatewaysEditCmd = &cobra.Command{
	Use:   "edit [GatewayID]",
	Short: "Edit a gateway",
	Long:  `ttnctl gateways edit can be used to edit settings of a gateway`,
	Example: `$ ttnctl gateways edit test --location 52.37403,4.88968 --frequency-plan EU
  INFO Edited gateway                          Gateway ID=test
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		gatewayID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(gatewayID, "Gateway ID"); err != nil {
			ctx.Fatal(err.Error())
		}

		var edits account.GatewayEdits

		if owner, _ := cmd.Flags().GetString("owner"); owner != "" {
			edits.Owner = owner
		}

		if frequencyPlan, _ := cmd.Flags().GetString("frequency-plan"); frequencyPlan != "" {
			edits.FrequencyPlan = frequencyPlan
		}

		if locationStr, _ := cmd.Flags().GetString("location"); locationStr != "" {
			location, err := util.ParseLocation(locationStr)
			if err != nil {
				ctx.WithError(err).Fatal("Invalid location")
			}
			edits.AntennaLocation = location
		}

		if public, _ := cmd.Flags().GetBool("location-public"); public {
			edits.LocationPublic = pointer.Bool(true)
		}
		if private, _ := cmd.Flags().GetBool("location-private"); private {
			edits.LocationPublic = pointer.Bool(false)
		}

		if public, _ := cmd.Flags().GetBool("status-public"); public {
			edits.StatusPublic = pointer.Bool(true)
		}
		if private, _ := cmd.Flags().GetBool("status-private"); private {
			edits.StatusPublic = pointer.Bool(false)
		}

		if public, _ := cmd.Flags().GetBool("owner-public"); public {
			edits.OwnerPublic = pointer.Bool(true)
		}
		if private, _ := cmd.Flags().GetBool("owner-private"); private {
			edits.OwnerPublic = pointer.Bool(false)
		}

		if router, _ := cmd.Flags().GetString("router"); router != "" {
			edits.Router = &router
		}

		var (
			attrs       = new(account.GatewayAttributes)
			attrsEdited bool
		)

		if brand, _ := cmd.Flags().GetString("brand"); brand != "" {
			attrs.Brand, attrsEdited = &brand, true
		}
		if model, _ := cmd.Flags().GetString("model"); model != "" {
			attrs.Model, attrsEdited = &model, true
		}
		if antennaType, _ := cmd.Flags().GetString("antenna-type"); antennaType != "" {
			attrs.AntennaType, attrsEdited = &antennaType, true
		}
		if antennaModel, _ := cmd.Flags().GetString("antenna-model"); antennaModel != "" {
			attrs.AntennaModel, attrsEdited = &antennaModel, true
		}
		if description, _ := cmd.Flags().GetString("description"); description != "" {
			attrs.Description, attrsEdited = &description, true
		}

		if attrsEdited {
			edits.Attributes = attrs
		}

		act := util.GetAccount(ctx)
		if err := act.EditGateway(gatewayID, edits); err != nil {
			ctx.WithError(err).Fatal("Failure editing gateway")
		}

		ctx.WithField("Gateway ID", gatewayID).Info("Edited gateway")
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysEditCmd)
	gatewaysEditCmd.Flags().String("owner", "", "The owner of the gateway")
	gatewaysEditCmd.Flags().String("frequency-plan", "", "The frequency plan to use on the gateway")
	gatewaysEditCmd.Flags().String("location", "", "The location of the gateway")

	gatewaysEditCmd.Flags().Bool("location-public", false, "Make the location of the gateway public")
	gatewaysEditCmd.Flags().Bool("location-private", false, "Make the location of the gateway private")
	gatewaysEditCmd.Flags().Bool("status-public", false, "Make the status of the gateway public")
	gatewaysEditCmd.Flags().Bool("status-private", false, "Make the status of the gateway private")
	gatewaysEditCmd.Flags().Bool("owner-public", false, "Make the owner of the gateway public")
	gatewaysEditCmd.Flags().Bool("owner-private", false, "Make the owner of the gateway private")

	gatewaysEditCmd.Flags().String("router", "", "The router of the gateway")

	gatewaysEditCmd.Flags().String("brand", "", "The brand of the gateway")
	gatewaysEditCmd.Flags().String("model", "", "The model of the gateway")
	gatewaysEditCmd.Flags().String("antenna-type", "", "The antenna type of the gateway")
	gatewaysEditCmd.Flags().String("antenna-model", "", "The antenna model of the gateway")
	gatewaysEditCmd.Flags().String("description", "", "The description of the gateway")
}
