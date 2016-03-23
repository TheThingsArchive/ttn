// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"crypto/aes"
	"encoding/base64"
	"net"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/brocaar/lorawan"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// joinCmd represents the `join-request` command
var joinCmd = &cobra.Command{
	Use:   "join [DevEUI] [AppKey]",
	Short: "Send a join requests to the network",
	Long:  `ttnctl join sends a join request to the network`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			ctx.Fatalf("Insufficient arguments")
		}

		// Parse parameters
		devEUIRaw, err := util.Parse64(args[0])
		if err != nil {
			ctx.Fatalf("Invalid DevEUI: %s", err)
		}
		var devEUI lorawan.EUI64
		copy(devEUI[:], devEUIRaw)

		appKeyRaw, err := util.Parse128(args[1])
		if err != nil {
			ctx.Fatalf("Invalid appKey: %s", err)
		}
		var appKey lorawan.AES128Key
		copy(appKey[:], appKeyRaw)

		appEUIRaw, err := util.Parse64(viper.GetString("join.app-eui"))
		if err != nil {
			ctx.Fatalf("Invalid appEUI: %s", err)
		}
		var appEUI lorawan.EUI64
		copy(appEUI[:], appEUIRaw)

		// Generate a DevNonce
		var devNonce [2]byte
		copy(devNonce[:], util.RandToken())

		// Lorawan Payload
		joinPayload := lorawan.JoinRequestPayload{
			AppEUI:   appEUI,
			DevEUI:   devEUI,
			DevNonce: devNonce,
		}
		phyPayload := lorawan.NewPHYPayload(true)
		phyPayload.MHDR = lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		}
		phyPayload.MACPayload = &joinPayload
		if err := phyPayload.SetMIC(appKey); err != nil {
			ctx.Fatalf("Unable to set MIC: %s", err)
		}

		addr, err := net.ResolveUDPAddr("udp", viper.GetString("router.address"))
		if err != nil {
			ctx.Fatalf("Couldn't resolve UDP address: %s", err)
		}
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			ctx.Fatalf("Couldn't Dial UDP connection: %s", err)
		}

		// Handle downlink
		chdown := make(chan bool)
		go func() {
			// Get Ack
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				ctx.Fatalf("Error receiving udp datagram: %s", err)
			}
			pkt := new(semtech.Packet)
			if err := pkt.UnmarshalBinary(buf[:n]); err != nil {
				ctx.Fatalf("Invalid udp response: %s", err)
			}
			ctx.Infof("Received Ack: %s", pkt)

			// Get Downlink, if any
			buf = make([]byte, 1024)
			n, err = conn.Read(buf)
			if err != nil {
				ctx.Fatalf("Error receiving udp datagram: %s", err)
			}
			pkt = new(semtech.Packet)
			if err = pkt.UnmarshalBinary(buf[:n]); err != nil {
				ctx.Fatalf("Invalid udp response: %s", err)
			}
			ctx.Infof("Received Downlink: %s", pkt)
			defer func() { chdown <- true }()

			if pkt.Payload == nil || pkt.Payload.TXPK == nil || pkt.Payload.TXPK.Data == nil {
				ctx.Fatalf("No payload available in downlink response")
			}

			data, err := base64.RawStdEncoding.DecodeString(*pkt.Payload.TXPK.Data)
			if err != nil {
				ctx.Fatalf("Unable to decode data payload: %s", err)
			}

			payload := lorawan.NewPHYPayload(false)
			if err := payload.UnmarshalBinary(data); err != nil {
				ctx.Fatalf("Unable to retrieve LoRaWAN PhyPayload: %s", err)
			}

			if err := payload.DecryptJoinAcceptPayload(appKey); err != nil {
				ctx.Fatalf("Unable to decrypt MACPayload: %s", err)
			}

			joinAccept, ok := payload.MACPayload.(*lorawan.JoinAcceptPayload)
			if !ok {
				ctx.Fatalf("Unable to retrieve LoRaWAN Join-Accept Payload")
			}

			buf = make([]byte, 16)
			copy(buf[1:4], joinAccept.AppNonce[:])
			copy(buf[4:7], joinAccept.NetID[:])
			copy(buf[7:9], devNonce[:])

			block, err := aes.NewCipher(appKey[:])
			if err != nil || block.BlockSize() != 16 {
				ctx.Fatalf("Unable to create cipher to generate keys: %s", err)
			}

			var nwkSKey, appSKey [16]byte
			buf[0] = 0x1
			block.Encrypt(nwkSKey[:], buf)
			buf[0] = 0x2
			block.Encrypt(appSKey[:], buf)

			ctx.Info("Network Joined.")
			ctx.Infof("Device Address: %X", joinAccept.DevAddr[:])
			ctx.Infof("Network Session Key: %X", nwkSKey)
			ctx.Infof("Application Session Key: %X", appSKey)
			ctx.Infof("Available Frequencies: %v", joinAccept.CFList)
		}()

		// Router Packet
		data, err := phyPayload.MarshalBinary()
		if err != nil {
			ctx.Fatalf("Couldn't construct LoRaWAN physical payload: %s", err)
		}
		encoded := strings.Trim(base64.StdEncoding.EncodeToString(data), "=")
		payload := semtech.Packet{
			Identifier: semtech.PUSH_DATA,
			Token:      util.RandToken(),
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Version:    semtech.VERSION,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Rssi: pointer.Int32(util.RandRssi()),
						Lsnr: pointer.Float32(util.RandLsnr()),
						Freq: pointer.Float32(util.RandFreq()),
						Datr: pointer.String(util.RandDatr()),
						Codr: pointer.String(util.RandCodr()),
						Modu: pointer.String("LoRa"),
						Tmst: pointer.Uint32(1),
						Data: &encoded,
					},
				},
			},
		}

		ctx.Infof("Sending packet: %s", payload.String())

		data, err = payload.MarshalBinary()
		if err != nil {
			ctx.Fatalf("Unable to construct framepayload: %v", data)
		}

		_, err = conn.Write(data)
		if err != nil {
			ctx.Fatal("Unable to send payload")
		}

		select {
		case <-chdown:
		case <-time.After(2 * time.Second):
		}
	},
}

func init() {
	RootCmd.AddCommand(joinCmd)

	joinCmd.Flags().String("ttn-router", "0.0.0.0:1700", "The net address of the TTN Router")
	viper.BindPFlag("router.address", joinCmd.Flags().Lookup("ttn-router"))

	joinCmd.Flags().String("app-eui", "0102030405060708", "The app EUI to use")
	viper.BindPFlag("join.app-eui", joinCmd.Flags().Lookup("app-eui"))
}
