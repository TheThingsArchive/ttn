// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsAddCmd = &cobra.Command{
	Use:   "add [AppID] [Description]",
	Short: "Add a new application",
	Long:  `ttnctl applications add can be used to add a new application to your account.`,
	Example: `$ ttnctl applications add test "Test application"
  INFO Added Application
  INFO Selected Current Application
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 2, 2)

		var euis []types.AppEUI
		euiStrings, err := cmd.Flags().GetStringSlice("app-eui")
		if err != nil {
			ctx.WithError(err).Fatal("Could not read app-eui options")
		}
		for _, euiString := range euiStrings {
			eui, err := types.ParseAppEUI(euiString)
			if err != nil {
				ctx.WithError(err).Fatal("Could not read app-eui")
			}
			euis = append(euis, eui)
		}

		account := util.GetAccount(ctx)

		app, err := account.CreateApplication(args[0], args[1], euis)
		if err != nil {
			ctx.WithError(err).Fatal("Could not add application")
		}

		util.ForceRefreshToken(ctx)

		ctx.Info("Added Application")

		skipSelect, _ := cmd.Flags().GetBool("skip-select")
		if !skipSelect {
			util.SetApp(ctx, app.ID, app.EUIs[0])
		}

		ctx.Info("Selected Current Application")

	},
}

func init() {
	applicationsCmd.AddCommand(applicationsAddCmd)
	applicationsAddCmd.Flags().StringSlice("app-eui", []string{}, "LoRaWAN AppEUI to register with application")
	applicationsAddCmd.Flags().Bool("skip-select", false, "Do not select this application (also adds --skip-register)")
	applicationsAddCmd.Flags().Bool("skip-register", false, "Do not register application with the Handler")
}
