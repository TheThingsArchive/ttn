// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/api/router"
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
		ctx.WithField("Gateway EUI", euiString)

		eui, err := types.ParseGatewayEUI(euiString)
		if err != nil {
			ctx.WithError(err).Fatal("Invalid Gateway EUI")
		}

		conn, manager := util.GetRouterManager(ctx)
		defer conn.Close()

		st, err := manager.GatewayStatus(util.GetContext(ctx), &router.GatewayStatusRequest{
			GatewayEui: &eui,
		})
		if err != nil {
			ctx.WithError(err).Fatal("Could not get status of gateway.")
		}

		ctx.Infof("Status of gateway %s:", eui)
		fmt.Println()
		printKV("Last seen", time.Unix(0, st.LastSeen))
		printKV("Timestamp", time.Duration(st.Status.Timestamp))
		if t := st.Status.Time; t != 0 {
			printKV("Reported time", time.Unix(0, t))
		}
		printKV("Description", st.Status.Description)
		printKV("Platform", st.Status.Platform)
		printKV("Contact email", st.Status.ContactEmail)
		printKV("Region", st.Status.Region)
		printKV("GPS coordinates", func() interface{} {
			if gps := st.Status.Gps; gps != nil || gps.Latitude != 0 && gps.Longitude != 0 {
				return fmt.Sprintf("(%v %v %v)", gps.Latitude, gps.Longitude, gps.Altitude)
			}
			return "not available"
		}())
		printKV("Rtt", func() interface{} {
			if t := st.Status.Rtt; t != 0 {
				return time.Duration(t)
			}
			return "unknown"
		}())
		printKV("Rx", fmt.Sprintf("(in: %v; ok: %v)", st.Status.RxIn, st.Status.RxOk))
		printKV("Tx", fmt.Sprintf("(in: %v; ok: %v)", st.Status.TxIn, st.Status.TxOk))
		fmt.Println()
	},
}

func init() {
	gatewayCmd.AddCommand(gatewayStatusCmd)
}
