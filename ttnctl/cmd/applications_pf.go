// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsPayloadFunctionsCmd = &cobra.Command{
	Use:   "pf",
	Short: "Show the payload functions",
	Long: `ttnctl applications pf shows the payload functions for decoding,
converting and validating binary payload.`,
	Example: `$ ttnctl applications pf
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Found Application
  INFO Decoder function
function Decoder(bytes, port) {
  var decoded = {};
  if (port === 1) {
    decoded.led = bytes[0];
  }
  return decoded;
}
  INFO No converter function
  INFO No validator function
  INFO No encoder function
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 0, 0)

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		app, err := manager.GetApplication(appID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get application.")
		}

		ctx.Info("Found Application")

		if app.Decoder != "" {
			ctx.Info("Decoder function")
			fmt.Println(app.Decoder)
		} else {
			ctx.Info("No decoder function")
		}

		if app.Converter != "" {
			ctx.Info("Converter function")
			fmt.Println(app.Converter)
		} else {
			ctx.Info("No converter function")
		}

		if app.Validator != "" {
			ctx.Info("Validator function")
			fmt.Println(app.Validator)
		} else {
			ctx.Info("No validator function")
		}

		if app.Encoder != "" {
			ctx.Info("Encoder function")
			fmt.Println(app.Encoder)
		} else {
			ctx.Info("No encoder function")
		}
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsPayloadFunctionsCmd)
}
