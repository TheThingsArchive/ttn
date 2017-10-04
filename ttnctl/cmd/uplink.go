// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"strconv"
	"time"

	"github.com/TheThingsNetwork/api/router"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var uplinkCmd = &cobra.Command{
	Hidden: true,
	Use:    "uplink [DevAddr] [NwkSKey] [AppSKey] [FCnt] [Payload]",
	Short:  "Simulate an uplink message to the network",
	Long:   `ttnctl uplink simulates an uplink message to the network`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 4, 5)

		devAddr, err := types.ParseDevAddr(args[0])
		if err != nil {
			ctx.WithError(err).Fatal("Invalid DevAddr")
		}

		nwkSKey, err := types.ParseNwkSKey(args[1])
		if err != nil {
			ctx.WithError(err).Fatal("Invalid NwkSKey")
		}

		appSKey, err := types.ParseAppSKey(args[2])
		if err != nil {
			ctx.WithError(err).Fatal("Invalid AppSKey")
		}

		fCnt, err := strconv.Atoi(args[3])
		if err != nil {
			ctx.WithError(err).Fatal("Invalid FCnt")
		}

		var payload []byte
		if len(args) >= 5 {
			payload, err = types.ParseHEX(args[4], len(args[4])/2)
			if err != nil {
				ctx.WithError(err).Fatal("Invalid Payload")
			}
		}

		confirmed, _ := cmd.Flags().GetBool("confirmed")

		ack, _ := cmd.Flags().GetBool("ack")

		rtrConn, rtrClient := util.GetRouter(ctx)
		defer rtrConn.Close()
		defer rtrClient.Close()

		gatewayID := viper.GetString("gateway-id")
		gatewayToken := viper.GetString("gateway-token")

		if gatewayID != "dev" {
			account := util.GetAccount(ctx)
			token, err := account.GetGatewayToken(gatewayID)
			if err != nil {
				ctx.WithError(err).Warn("Could not get gateway token")
				ctx.Warn("Trying without token. Your message may not be processed by the router")
				gatewayToken = ""
			} else if token != nil && token.AccessToken != "" {
				gatewayToken = token.AccessToken
			}
		}

		gtwClient := rtrClient.NewGatewayStreams(gatewayID, gatewayToken, true)
		defer gtwClient.Close()

		time.Sleep(100 * time.Millisecond)

		m := &util.Message{}
		m.SetDevice(devAddr, nwkSKey, appSKey)
		m.SetMessage(confirmed, ack, fCnt, payload)
		bytes := m.Bytes()

		gtwClient.Uplink(&router.UplinkMessage{
			Payload:          bytes,
			GatewayMetadata:  *util.GetGatewayMetadata(gatewayID, 868100000),
			ProtocolMetadata: *util.GetProtocolMetadata("SF7BW125"),
		})
		if err != nil {
			ctx.WithError(err).Fatal("Could not send uplink to Router")
		}

		time.Sleep(100 * time.Millisecond)

		ctx.Info("Sent uplink to Router")

		downlink, _ := gtwClient.Downlink()
		select {
		case downlinkMessage, ok := <-downlink:
			if !ok {
				ctx.Info("Did not receive downlink")
				break
			}
			if err := m.Unmarshal(downlinkMessage.Payload); err != nil {
				ctx.WithError(err).Fatal("Could not unmarshal downlink")
			}
			ctx.WithFields(ttnlog.Fields{
				"Payload": m.Payload,
				"FCnt":    m.FCnt,
				"FPort":   m.FPort,
			}).Info("Received Downlink")
		case <-time.After(2 * time.Second):
			ctx.Info("Did not receive downlink")
		}
	},
}

func init() {
	RootCmd.AddCommand(uplinkCmd)
	uplinkCmd.Flags().Bool("confirmed", false, "Use confirmed uplink")
	uplinkCmd.Flags().Bool("ack", false, "Set ACK bit")

	uplinkCmd.Flags().String("gateway-id", "", "The ID of the gateway that you are faking (you can only fake gateways that you own)")
	viper.BindPFlag("gateway-id", uplinkCmd.Flags().Lookup("gateway-id"))
}
