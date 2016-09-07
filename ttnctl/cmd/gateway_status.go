// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var gatewayStatusCmd = &cobra.Command{
	Use:   "status [GatewayEUI]",
	Short: "Get status of a gateway",
	Long:  `ttnctl gateway status can be used to get status of gateways.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return
		}

		euiString := args[0]
		ctx = ctx.WithField("Gateway EUI", euiString)

		eui, err := types.ParseGatewayEUI(euiString)
		if err != nil {
			ctx.WithError(err).Fatal("Invalid Gateway EUI")
		}

		conn, manager := util.GetRouterManager(ctx)
		defer conn.Close()

		resp, err := manager.GatewayStatus(util.GetContext(ctx), &router.GatewayStatusRequest{
			GatewayEui: &eui,
		})
		if err != nil {
			ctx.WithError(core.FromGRPCError(err)).Fatal("Could not get status of gateway.")
		}

		ctx.Infof("Received status")
		fmt.Println()
		printKV("Last seen", time.Unix(0, resp.LastSeen))
		printKV("Timestamp", time.Duration(resp.Status.Timestamp))
		if t := resp.Status.Time; t != 0 {
			printKV("Reported time", time.Unix(0, t))
		}
		printKV("Description", resp.Status.Description)
		printKV("Platform", resp.Status.Platform)
		printKV("Contact email", resp.Status.ContactEmail)
		printKV("Region", resp.Status.Region)
		printKV("GPS coordinates", func() interface{} {
			if gps := resp.Status.Gps; gps != nil && !(gps.Latitude == 0 && gps.Longitude == 0) {
				return fmt.Sprintf("(%.6f %.6f)", gps.Latitude, gps.Longitude)
			}
			return "not available"
		}())
		printKV("Rtt", func() interface{} {
			if t := resp.Status.Rtt; t != 0 {
				return time.Duration(t)
			}
			return "not available"
		}())
		printKV("Rx", fmt.Sprintf("(in: %d; ok: %d)", resp.Status.RxIn, resp.Status.RxOk))
		printKV("Tx", fmt.Sprintf("(in: %d; ok: %d)", resp.Status.TxIn, resp.Status.TxOk))
		fmt.Println()
	},
}

func init() {
	gatewayCmd.AddCommand(gatewayStatusCmd)
}
