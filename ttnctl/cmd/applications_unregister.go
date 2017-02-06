// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsUnregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister this application from the handler",
	Long:  `ttnctl unregister can be used to unregister this application from the handler.`,
	Example: `$ ttnctl applications unregister
Are you sure you want to unregister application test?
> yes
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Unregistered application                 AppID=test
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 0, 0)

		appID := util.GetAppID(ctx)

		if !confirm(fmt.Sprintf("Are you sure you want to unregister application %s?", appID)) {
			ctx.Info("Not doing anything")
			return
		}

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		err := manager.DeleteApplication(appID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not unregister application")
		}

		ctx.WithFields(ttnlog.Fields{
			"AppID": appID,
		}).Infof("Unregistered application")
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsUnregisterCmd)
}
