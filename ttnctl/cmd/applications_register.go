// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register this application with the handler",
	Long:  `ttnctl applications register can be used to register this application with the handler.`,
	Example: `$ ttnctl applications register
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Registered application                   AppID=test
`,
	Run: func(cmd *cobra.Command, args []string) {

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		err := manager.RegisterApplication(appID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not register application")
		}

		ctx.WithFields(ttnlog.Fields{
			"AppID": appID,
		}).Infof("Registered application")
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsRegisterCmd)
}
