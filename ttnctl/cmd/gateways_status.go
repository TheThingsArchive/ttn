// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/spf13/cobra"
)

var gatewaysStatusCmd = &cobra.Command{
	Use:    "status [gatewayID]",
	Short:  "Get status of a gateway",
	Long:   `ttnctl gateways status can be used to get status of gateways.`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 1)

		gatewayID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(gatewayID, "Gateway ID"); err != nil {
			ctx.Fatal(err.Error())
		}

		conn, manager := util.GetRouterManager(ctx)
		defer conn.Close()

		ctx = ctx.WithField("GatewayID", gatewayID)

		resp, err := manager.GatewayStatus(util.GetContext(ctx), &router.GatewayStatusRequest{
			GatewayId: gatewayID,
		})
		if err != nil {
			ctx.WithError(errors.FromGRPCError(err)).Fatal("Could not get status of gateway.")
		}

		ctx.Infof("Received status")
		fmt.Println()
		printKV("Last seen", time.Unix(0, resp.LastSeen))
		printKV("Timestamp", resp.Status.Timestamp)
		if t := resp.Status.Time; t != 0 {
			printKV("Reported time", time.Unix(0, t))
		}
		printKV("Description", resp.Status.Description)
		printKV("Platform", resp.Status.Platform)
		printKV("Contact email", resp.Status.ContactEmail)
		printKV("Frequency Plan", resp.Status.FrequencyPlan)
		printKV("Bridge", resp.Status.Bridge)
		printKV("IP Address", strings.Join(resp.Status.Ip, ", "))
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
		if resp.Uplink != nil {
			printKV("RxRate", fmt.Sprintf("%.4f, %.4f, %.4f", resp.Uplink.Rate1, resp.Uplink.Rate5, resp.Uplink.Rate15))
		}
		printKV("Tx", fmt.Sprintf("(in: %d; ok: %d)", resp.Status.TxIn, resp.Status.TxOk))
		if resp.Downlink != nil {
			printKV("TxRate", fmt.Sprintf("%.4f, %.4f, %.4f", resp.Downlink.Rate1, resp.Downlink.Rate5, resp.Downlink.Rate15))
		}
		fmt.Println()
	},
}

func init() {
	gatewaysCmd.AddCommand(gatewaysStatusCmd)
}
