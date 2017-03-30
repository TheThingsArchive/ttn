// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/otaa"
	"github.com/brocaar/lorawan"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var joinCmd = &cobra.Command{
	Hidden: true,
	Use:    "join [AppEUI] [DevEUI] [AppKey] [DevNonce]",
	Short:  "Simulate an join message to the network",
	Long:   `ttnctl join simulates an join message to the network`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 4, 4)

		appEUI, err := types.ParseAppEUI(args[0])
		if err != nil {
			ctx.WithError(err).Fatal("Could not parse AppEUI")
		}

		devEUI, err := types.ParseDevEUI(args[1])
		if err != nil {
			ctx.WithError(err).Fatal("Could not parse DevEUI")
		}

		appKey, err := types.ParseAppKey(args[2])
		if err != nil {
			ctx.WithError(err).Fatal("Could not parse AppKey")
		}

		devNonceSlice, err := types.ParseHEX(args[3], 2)
		if err != nil {
			ctx.WithError(err).Fatal("Could not parse DevNonce")
		}
		devNonce := [2]byte{devNonceSlice[0], devNonceSlice[1]}

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

		joinReq := &pb_lorawan.Message{
			MHDR: pb_lorawan.MHDR{MType: pb_lorawan.MType_JOIN_REQUEST, Major: pb_lorawan.Major_LORAWAN_R1},
			Payload: &pb_lorawan.Message_JoinRequestPayload{JoinRequestPayload: &pb_lorawan.JoinRequestPayload{
				AppEui:   appEUI,
				DevEui:   devEUI,
				DevNonce: types.DevNonce(devNonce),
			}}}
		joinPhy := joinReq.PHYPayload()
		joinPhy.SetMIC(lorawan.AES128Key(appKey))
		bytes, _ := joinPhy.MarshalBinary()

		uplink := &router.UplinkMessage{
			Payload:          bytes,
			GatewayMetadata:  util.GetGatewayMetadata(gatewayID, 868100000),
			ProtocolMetadata: util.GetProtocolMetadata("SF7BW125"),
		}
		uplink.UnmarshalPayload()

		gtwClient.Uplink(uplink)

		time.Sleep(100 * time.Millisecond)

		ctx.Info("Sent uplink to Router")

		downlink, _ := gtwClient.Downlink()
		select {
		case downlinkMessage, ok := <-downlink:
			if !ok {
				ctx.Info("Did not receive downlink")
				break
			}
			downlinkMessage.UnmarshalPayload()
			resPhy := downlinkMessage.Message.GetLorawan().PHYPayload()
			resPhy.DecryptJoinAcceptPayload(lorawan.AES128Key(appKey))
			res := pb_lorawan.MessageFromPHYPayload(resPhy)
			accept := res.GetJoinAcceptPayload()

			appSKey, nwkSKey, _ := otaa.CalculateSessionKeys(appKey, accept.AppNonce, accept.NetId, devNonce)

			ctx.WithFields(ttnlog.Fields{
				"DevAddr": accept.DevAddr,
				"NwkSKey": nwkSKey,
				"AppSKey": appSKey,
			}).Info("Received JoinAccept")
		case <-time.After(6 * time.Second):
			ctx.Info("Did not receive downlink")
		}

	},
}

func init() {
	RootCmd.AddCommand(joinCmd)

	joinCmd.Flags().String("gateway-id", "", "The ID of the gateway that you are faking (you can only fake gateways that you own)")
	viper.BindPFlag("gateway-id", joinCmd.Flags().Lookup("gateway-id"))
}
