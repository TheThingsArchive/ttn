// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

// applicationsCreateCmd is the entrypoint for handlerctl
var applicationsCreateCmd = &cobra.Command{
	Use:   "create [AppID] [Description]",
	Short: "Create a new application",
	Long:  `ttnctl applications create can be used to create a new application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.UsageFunc()(cmd)
			return
		}

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
			ctx.WithError(err).Fatal("Could not create application")
		}

		util.ForceRefreshToken(ctx)

		ctx.Info("Created Application")

		skipSelect, _ := cmd.Flags().GetBool("skip-select")
		if !skipSelect {
			err = util.SetConfig(map[string]interface{}{
				"app-id":  app.ID,
				"app-eui": app.EUIs[0].String(),
			})
			if err != nil {
				ctx.WithError(err).Fatal("Could not update configuration")
			}
		}

		ctx.Info("Selected Current Application")

	},
}

func init() {
	applicationsCmd.AddCommand(applicationsCreateCmd)
	applicationsCreateCmd.Flags().StringSlice("app-eui", []string{}, "LoRaWAN AppEUI to register with application")
	applicationsCreateCmd.Flags().Bool("skip-select", false, "Do not select this application (also adds --skip-register)")
	applicationsCreateCmd.Flags().Bool("skip-register", false, "Do not register application with the Handler")
}
