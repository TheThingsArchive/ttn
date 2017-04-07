// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsPayloadFormatCmd = &cobra.Command{
	Use:   "pf",
	Short: "Show the payload format",
	Long: `ttnctl applications pf shows the payload format to handle
binary payload.`,
	Example: `$ ttnctl applications pf
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Found application
  INFO Custom decoder function
function Decoder(bytes, port) {
  var decoded = {};
  if (port === 1) {
    decoded.led = bytes[0];
  }
  return decoded;
}
  INFO No custom converter function
  INFO No custom validator function
  INFO No custom encoder function
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

		ctx.Info("Found application")

		switch app.PayloadFormat {
		case "custom":
			if app.Decoder != "" {
				ctx.Info("Custom decoder function")
				fmt.Println(app.Decoder)
			} else {
				ctx.Info("No custom decoder function")
			}

			if app.Converter != "" {
				ctx.Info("Custom converter function")
				fmt.Println(app.Converter)
			} else {
				ctx.Info("No custom converter function")
			}

			if app.Validator != "" {
				ctx.Info("Custom validator function")
				fmt.Println(app.Validator)
			} else {
				ctx.Info("No custom validator function")
			}

			if app.Encoder != "" {
				ctx.Info("Custom encoder function")
				fmt.Println(app.Encoder)
			} else {
				ctx.Info("No custom encoder function")
			}
		default:
			ctx.Infof("Payload format set to %s", app.PayloadFormat)
		}
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsPayloadFormatCmd)
}
