// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var devicesInfoCmd = &cobra.Command{
	Use:   "info [Device ID]",
	Short: "Get information about a device",
	Long:  `ttnctl devices info can be used to get information about a device.`,
	Example: `$ ttnctl devices info test
  INFO Using Application                        AppEUI=70B3D57EF0000024 AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Found device

  Application ID: test
       Device ID: test
       Last Seen: never

    LoRaWAN Info:

     AppEUI: 70B3D57EF0000024
     DevEUI: 0001D544B2936FCE
    DevAddr: 26001ADA
     AppKey: <nil>
    AppSKey: D8DD37B4B709BA76C6FEC62CAD0CCE51
    NwkSKey: 3382A3066850293421ED8D392B9BF4DF
     FCntUp: 0
   FCntDown: 0
    Options:
`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			cmd.UsageFunc()(cmd)
			return
		}

		devID := args[0]
		if !api.ValidID(devID) {
			ctx.Fatalf("Invalid Device ID") // TODO: Add link to wiki explaining device IDs
		}

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx)
		defer conn.Close()

		dev, err := manager.GetDevice(appID, devID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get existing device.")
		}

		byteFormat, _ := cmd.Flags().GetString("format")

		ctx.Info("Found device")

		fmt.Println()

		fmt.Printf("  Application ID: %s\n", dev.AppId)
		fmt.Printf("       Device ID: %s\n", dev.DevId)
		if lorawan := dev.GetLorawanDevice(); lorawan != nil {
			lastSeen := "never"
			if lorawan.LastSeen > 0 {
				lastSeen = fmt.Sprintf("%s", time.Unix(0, 0).Add(time.Duration(lorawan.LastSeen)))
			}

			fmt.Printf("       Last Seen: %s\n", lastSeen)
			fmt.Println()
			fmt.Println("    LoRaWAN Info:")
			fmt.Println()
			fmt.Printf("     AppEUI: %s\n", formatBytes(lorawan.AppEui, byteFormat))
			fmt.Printf("     DevEUI: %s\n", formatBytes(lorawan.DevEui, byteFormat))
			fmt.Printf("    DevAddr: %s\n", formatBytes(lorawan.DevAddr, byteFormat))
			fmt.Printf("     AppKey: %s\n", formatBytes(lorawan.AppKey, byteFormat))
			fmt.Printf("    AppSKey: %s\n", formatBytes(lorawan.AppSKey, byteFormat))
			fmt.Printf("    NwkSKey: %s\n", formatBytes(lorawan.NwkSKey, byteFormat))

			fmt.Printf("     FCntUp: %d\n", lorawan.FCntUp)
			fmt.Printf("   FCntDown: %d\n", lorawan.FCntDown)
			options := []string{}
			if lorawan.DisableFCntCheck {
				options = append(options, "DisableFCntCheck")
			}
			if lorawan.Uses32BitFCnt {
				options = append(options, "Uses32BitFCnt")
			}
			fmt.Printf("    Options: %s\n", strings.Join(options, ", "))
		}

	},
}

type formattableBytes interface {
	IsEmpty() bool
	Bytes() []byte
}

func formatBytes(toPrint interface{}, format string) string {
	if i, ok := toPrint.(formattableBytes); ok {
		if i.IsEmpty() {
			return "<nil>"
		}
		switch format {
		case "msb":
			return cStyle(i.Bytes(), true) + " (msb first)"
		case "lsb":
			return cStyle(i.Bytes(), false) + " (lsb first)"
		case "hex":
			return fmt.Sprintf("%X", i.Bytes())
		}
	}
	return fmt.Sprintf("%s", toPrint)
}

// cStyle prints the byte slice in C-Style
func cStyle(bytes []byte, msbf bool) string {
	output := "{"
	if !msbf {
		bytes = reverse(bytes)
	}
	for i, b := range bytes {
		if i != 0 {
			output += ", "
		}
		output += fmt.Sprintf("0x%02X", b)
	}
	output += "}"
	return output
}

// reverse is used to convert between MSB-first and LSB-first
func reverse(in []byte) (out []byte) {
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return
}

func init() {
	devicesCmd.AddCommand(devicesInfoCmd)
	devicesInfoCmd.Flags().String("format", "hex", "Formatting: hex/msb/lsb")
}
