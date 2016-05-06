// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"encoding/base64"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/brocaar/lorawan"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// uplinkCmd represents the `uplink` command
var uplinkCmd = &cobra.Command{
	Use:   "uplink [ShouldConfirm] [DevAddr] [NwkSKey] [AppSKey] [Payload] [FCnt]",
	Short: "Send uplink messages to the network",
	Long:  `ttnctl uplink sends an uplink message to the network`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 6 {
			ctx.Fatalf("Insufficient arguments")
		}

		// Parse parameters
		var mtype lorawan.MType
		switch args[0] {
		case "yes":
			fallthrough
		case "true":
			mtype = lorawan.ConfirmedDataUp
		default:
			mtype = lorawan.UnconfirmedDataUp
		}

		devAddr, err := types.ParseDevAddr(args[1])
		if err != nil {
			ctx.Fatalf("Invalid DevAddr: %s", err)
		}

		nwkSKey, err := types.ParseNwkSKey(args[2])
		if err != nil {
			ctx.Fatalf("Invalid NwkSKey: %s", err)
		}

		appSKey, err := types.ParseAppSKey(args[3])
		if err != nil {
			ctx.Fatalf("Invalid appSKey: %s", err)
		}

		fcnt, err := strconv.ParseInt(args[5], 10, 64)
		if err != nil {
			ctx.Fatalf("Invalid FCnt: %s", err)
		}

		// Lorawan Payload
		macPayload := &lorawan.MACPayload{}
		macPayload.FHDR = lorawan.FHDR{
			DevAddr: lorawan.DevAddr(devAddr),
			FCnt:    uint32(fcnt),
		}
		macPayload.FPort = pointer.Uint8(1)
		if plain, _ := cmd.Flags().GetBool("plain"); plain {
			ctx.Warn("Sending data as plain text is bad practice. We recommend to transmit data in a binary format.")
			macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(args[4])}}
		} else {
			payload, err := util.ParseHEX(args[4], len(args[4]))
			if err != nil {
				ctx.Fatalf("Invalid hexadecimal payload. If you are trying to send a plain-text payload, use the --plain flag.")
			}
			macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: payload}}
		}
		phyPayload := &lorawan.PHYPayload{}
		phyPayload.MHDR = lorawan.MHDR{
			MType: mtype,
			Major: lorawan.LoRaWANR1,
		}
		phyPayload.MACPayload = macPayload
		if err := phyPayload.EncryptFRMPayload(lorawan.AES128Key(appSKey)); err != nil {
			ctx.Fatalf("Unable to encrypt frame payload: %s", err)
		}
		if err := phyPayload.SetMIC(lorawan.AES128Key(nwkSKey)); err != nil {
			ctx.Fatalf("Unable to set MIC: %s", err)
		}

		addr, err := net.ResolveUDPAddr("udp", viper.GetString("ttn-router"))
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
			// Get PullAck
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				ctx.Fatalf("Error receiving udp datagram: %s", err)
			}
			pkt := new(semtech.Packet)
			if err := pkt.UnmarshalBinary(buf[:n]); err != nil {
				ctx.Fatalf("Invalid udp response: %s", err)
			}
			ctx.Infof("Received PullAck: %s", pkt)

			// Get Ack
			buf = make([]byte, 1024)
			n, err = conn.Read(buf)
			if err != nil {
				ctx.Fatalf("Error receiving udp datagram: %s", err)
			}
			pkt = new(semtech.Packet)
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

			payload := &lorawan.PHYPayload{}
			if err := payload.UnmarshalBinary(data); err != nil {
				ctx.Fatalf("Unable to retrieve LoRaWAN PhyPayload: %s", err)
			}

			micOK, _ := payload.ValidateMIC(lorawan.AES128Key(nwkSKey))
			if !micOK {
				ctx.Warn("MIC check failed.")
			}

			macPayload, ok := payload.MACPayload.(*lorawan.MACPayload)
			if !ok || len(macPayload.FRMPayload) > 1 {
				ctx.Fatalf("Unable to retrieve LoRaWAN MACPayload")
			}
			ctx.Infof("Frame counter: %d", macPayload.FHDR.FCnt)
			if len(macPayload.FRMPayload) > 0 {
				decrypted, err := lorawan.EncryptFRMPayload(
					lorawan.AES128Key(appSKey),
					false,
					lorawan.DevAddr(devAddr),
					macPayload.FHDR.FCnt,
					macPayload.FRMPayload[0].(*lorawan.DataPayload).Bytes,
				)
				if err != nil {
					ctx.Fatalf("Unable to decrypt MACPayload: %s", err)
				}
				if plain, _ := cmd.Flags().GetBool("plain"); plain {
					unprintable, _ := regexp.Compile(`[^[:print:]]`)
					if unprintable.Match(decrypted) {
						ctx.WithField("warning", "payload contains unprintable characters").Infof("Decrypted Payload: %X", decrypted)
					} else {
						ctx.Infof("%s", decrypted)
					}
				} else {
					ctx.Infof("Decrypted Payload: %X", decrypted)
				}
			} else {
				ctx.Infof("The frame payload was empty.")
			}
		}()

		// PULL_DATA Packet

		pullPacket := semtech.Packet{
			Version:    semtech.VERSION,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Token:      []byte{1, 2},
			Identifier: semtech.PULL_DATA,
		}
		pullData, err := pullPacket.MarshalBinary()
		if err != nil {
			ctx.Fatal("Unable to construct pull_data")
		}

		// Router Packet
		data, err := phyPayload.MarshalBinary()
		if err != nil {
			ctx.Fatalf("Couldn't construct LoRaWAN physical payload: %s", err)
		}
		encoded := strings.Trim(base64.StdEncoding.EncodeToString(data), "=")
		payload := semtech.Packet{
			Identifier: semtech.PUSH_DATA,
			Token:      random.Token(),
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Version:    semtech.VERSION,
			Payload: &semtech.Payload{
				RXPK: []semtech.RXPK{
					{
						Rssi: pointer.Int32(random.Rssi()),
						Lsnr: pointer.Float32(random.Lsnr()),
						Freq: pointer.Float32(random.Freq()),
						Datr: pointer.String(random.Datr()),
						Codr: pointer.String(random.Codr()),
						Modu: pointer.String("LoRa"),
						Tmst: pointer.Uint32(1),
						Time: pointer.Time(time.Now().UTC()),
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

		_, err = conn.Write(pullData)
		if err != nil {
			ctx.Fatal("Unable to send pull_data")
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
	RootCmd.AddCommand(uplinkCmd)
	uplinkCmd.Flags().Bool("plain", false, "send payload as plain-text")
}
