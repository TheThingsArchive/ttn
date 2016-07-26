// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

// applicationsPayloadFunctionsCmd represents the applicationsPayloadFunctions command
var applicationsPayloadFunctionsCmd = &cobra.Command{
	Use:   "pf",
	Short: "Show the payload functions",
	Long: `ttnctl applications pf shows the payload functions for decoding,
converting and validating binary payload.
`,
	Run: func(cmd *cobra.Command, args []string) {

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx)
		defer conn.Close()

		app, err := manager.GetApplication(appID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get application.")
		}

		ctx.Info("Found Application")

		if app.Decoder != "" {
			ctx.Info("Decoder function")
			fmt.Println(app.Decoder)
		}

		if app.Converter != "" {
			ctx.Info("Converter function")
			fmt.Println(app.Converter)
		}

		if app.Validator != "" {
			ctx.Info("Validator function")
			fmt.Println(app.Validator)
		}
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsPayloadFunctionsCmd)
	applicationsPayloadFunctionsCmd.AddCommand(applicationsPayloadFunctionsSetCmd)
}
