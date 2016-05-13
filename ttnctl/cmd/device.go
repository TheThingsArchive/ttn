// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/apex/log"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

const emptyCell = "-"

var defaultKey = []byte{0x2B, 0x7E, 0x15, 0x16, 0x28, 0xAE, 0xD2, 0xA6, 0xAB, 0xF7, 0x15, 0x88, 0x09, 0xCF, 0x4F, 0x3C}

// devicesCmd represents the `devices` command
var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage devices on the Handler",
	Long: `ttnctl devices retrieves a list of devices that your application
registered on the Handler.`,
	Run: func(cmd *cobra.Command, args []string) {

		appEUI := util.GetAppEUI(ctx)

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found. Please login")
		}

		manager := util.GetHandlerManager(ctx)
		defaultDevice, err := manager.GetDefaultDevice(context.Background(), &core.GetDefaultDeviceReq{
			Token:  auth.AccessToken,
			AppEUI: appEUI.Bytes(),
		})
		if err != nil {
			// TODO: Check reason; not found is OK here
			defaultDevice = nil
		}
		if defaultDevice != nil {
			ctx.Warn("Application activates new devices with default AppKey")
			fmt.Printf("Default AppKey:  %X\n", defaultDevice.AppKey)
			fmt.Printf("                 {%s}\n", cStyle(defaultDevice.AppKey, msbf))
		} else {
			ctx.Info("Application does not activate new devices with default AppKey")
		}

		devices, err := manager.ListDevices(context.Background(), &core.ListDevicesHandlerReq{
			Token:  auth.AccessToken,
			AppEUI: appEUI.Bytes(),
		})
		if err != nil {
			ctx.WithError(err).Fatal("Could not get device list")
		}

		ctx.Infof("Found %d personalized devices (ABP)", len(devices.ABP))

		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("DevAddr", "FCntUp", "FCntDown", "Flags")
		for _, device := range devices.ABP {
			devAddr := fmt.Sprintf("%X", device.DevAddr)
			var flags string
			if (device.Flags & core.RelaxFcntCheck) != 0 {
				flags = "relax-fcnt"
			}
			if flags == "" {
				flags = "-"
			}
			table.AddRow(devAddr, device.FCntUp, device.FCntDown, strings.TrimLeft(flags, ","))
		}

		fmt.Println()
		fmt.Println(table)
		fmt.Println()

		ctx.Infof("Found %d dynamic devices (OTAA)", len(devices.OTAA))
		table = uitable.New()
		table.MaxColWidth = 40
		table.AddRow("DevEUI", "DevAddr", "FCntUp", "FCntDown")
		for _, device := range devices.OTAA {
			devEUI := fmt.Sprintf("%X", device.DevEUI)
			devAddr := fmt.Sprintf("%X", device.DevAddr)
			table.AddRow(devEUI, devAddr, device.FCntUp, device.FCntDown)
		}

		fmt.Println()
		fmt.Println(table)
		fmt.Println()

		ctx.Info("Run 'ttnctl devices info [DevAddr|DevEUI]' for more information about a specific device")
	},
}

const (
	msbf = true
	lsbf = false
)

// devicesInfoCmd represents the `devices info` command
var devicesInfoCmd = &cobra.Command{
	Use:   "info [DevAddr|DevEUI]",
	Short: "Show device information",
	Long:  `ttnctl devices info shows information about a specific device.`,
	Run: func(cmd *cobra.Command, args []string) {
		appEUI := util.GetAppEUI(ctx)

		if len(args) != 1 {
			ctx.Fatal("Missing DevAddr or DevEUI")
		}

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found. Please login")
		}

		manager := util.GetHandlerManager(ctx)
		res, err := manager.ListDevices(context.Background(), &core.ListDevicesHandlerReq{
			Token:  auth.AccessToken,
			AppEUI: appEUI.Bytes(),
		})
		if err != nil {
			ctx.WithError(err).Fatal("Could not get device list")
		}

		lmic, _ := cmd.Flags().GetBool("lmic")

		if devEUI, err := types.ParseDevEUI(args[0]); err == nil {
			for _, device := range res.OTAA {
				if bytes.Equal(device.DevEUI, devEUI.Bytes()) {
					fmt.Println("Dynamic device:")

					fmt.Println()

					// LMiC decided to use LSBF for AppEUI and call it ArtEUI
					if lmic {
						fmt.Printf("  AppEUI:  %X (sometimes called ArtEUI)\n", appEUI)
						fmt.Printf("           {%s} (Note: LSBF)\n", cStyle(appEUI.Bytes(), lsbf))
					} else {
						fmt.Printf("  AppEUI:  %X\n", appEUI)
						fmt.Printf("           {%s}\n", cStyle(appEUI.Bytes(), msbf))
					}

					fmt.Println()
					fmt.Printf("  DevEUI:  %X\n", device.DevEUI)
					// LMiC decided to use LSBF for DevEUI
					if lmic {
						fmt.Printf("           {%s} (Note: LSBF)\n", cStyle(device.DevEUI, lsbf))
					} else {
						fmt.Printf("           {%s}\n", cStyle(device.DevEUI, msbf))
					}

					fmt.Println()
					// LMiC decided to rename AppKey to DevKey
					if lmic {
						fmt.Printf("  AppKey:  %X (sometimes called DevKey)\n", device.AppKey)
					} else {
						fmt.Printf("  AppKey:  %X\n", device.AppKey)
					}
					fmt.Printf("           {%s}\n", cStyle(device.AppKey, msbf))

					if len(device.DevAddr) != 0 {
						fmt.Println()
						fmt.Println("  Activated with the following parameters:")

						fmt.Println()
						fmt.Printf("  DevAddr: %X\n", device.DevAddr)
						fmt.Printf("           {%s}\n", cStyle(device.DevAddr, msbf))

						fmt.Println()
						fmt.Printf("  NwkSKey: %X\n", device.NwkSKey)
						fmt.Printf("           {%s}\n", cStyle(device.NwkSKey, msbf))

						fmt.Println()
						// LMiC decided to rename AppSKey to ArtSKey
						if lmic {
							fmt.Printf("  AppSKey:  %X (sometimes called ArtSKey)\n", device.AppSKey)
						} else {
							fmt.Printf("  AppSKey:  %X\n", device.AppSKey)
						}
						fmt.Printf("           {%s}\n", cStyle(device.AppSKey, msbf))

						fmt.Println()
						fmt.Printf("  FCntUp:  %d\n  FCntDn:  %d\n", device.FCntUp, device.FCntDown)
					} else {
						fmt.Println()
						fmt.Println("  Not yet activated")
					}

					return
				}
			}
		}

		if devAddr, err := types.ParseDevAddr(args[0]); err == nil {
			for _, device := range res.ABP {
				if bytes.Equal(device.DevAddr, devAddr.Bytes()) {
					fmt.Println("Personalized device:")

					fmt.Println()
					fmt.Printf("  DevAddr: %X\n", device.DevAddr)
					fmt.Printf("           {%s}\n", cStyle(device.DevAddr, msbf))

					fmt.Println()
					fmt.Printf("  NwkSKey: %X\n", device.NwkSKey)
					fmt.Printf("           {%s}\n", cStyle(device.NwkSKey, msbf))

					fmt.Println()
					// LMiC decided to rename AppSKey to ArtSKey
					if lmic {
						fmt.Printf("  AppSKey:  %X (sometimes called ArtSKey)\n", device.AppSKey)
					} else {
						fmt.Printf("  AppSKey:  %X\n", device.AppSKey)
					}
					fmt.Printf("           {%s}\n", cStyle(device.AppSKey, msbf))

					fmt.Println()
					fmt.Printf("  FCntUp:  %d\n  FCntDn:  %d\n", device.FCntUp, device.FCntDown)
					fmt.Println()
					var flags string
					if (device.Flags & core.RelaxFcntCheck) != 0 {
						flags = "relax-fcnt"
					}
					if flags == "" {
						flags = "-"
					}
					fmt.Printf("  Flags:   %s\n", strings.TrimLeft(flags, ","))
					return
				}
			}
		} else {
			ctx.Fatal("Invalid DevAddr or DevEUI")
		}

		ctx.Info("Device not found")

	},
}

func cStyle(bytes []byte, msbf bool) (output string) {
	if !msbf {
		bytes = reverse(bytes)
	}
	for i, b := range bytes {
		if i != 0 {
			output += ", "
		}
		output += fmt.Sprintf("0x%02X", b)
	}
	return
}

// reverse is used to convert between MSB-first and LSB-first
func reverse(in []byte) (out []byte) {
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return
}

// devicesRegisterCmd represents the `device register` command
var devicesRegisterCmd = &cobra.Command{
	Use:   "register [DevEUI] [AppKey]",
	Short: "Create or Update registrations on the Handler",
	Long: `ttnctl devices register creates or updates an OTAA registration on
the Handler`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		appEUI := util.GetAppEUI(ctx)

		devEUI, err := types.ParseDevEUI(args[0])
		if err != nil {
			ctx.Fatalf("Invalid DevEUI: %s", err)
		}

		var appKey types.AppKey
		if len(args) >= 2 {
			appKey, err = types.ParseAppKey(args[1])
			if err != nil {
				ctx.Fatalf("Invalid AppKey: %s", err)
			}
		} else {
			ctx.Info("Generating random AppKey...")
			copy(appKey[:], random.Bytes(16))
		}

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found. Please login")
		}

		manager := util.GetHandlerManager(ctx)
		res, err := manager.UpsertOTAA(context.Background(), &core.UpsertOTAAHandlerReq{
			Token:  auth.AccessToken,
			AppEUI: appEUI.Bytes(),
			DevEUI: devEUI.Bytes(),
			AppKey: appKey.Bytes(),
		})
		if err != nil || res == nil {
			ctx.WithError(err).Fatal("Could not register device")
		}
		ctx.WithFields(log.Fields{
			"DevEUI": devEUI,
			"AppKey": appKey,
		}).Info("Registered device")
	},
}

// devicesRegisterPersonalizedCmd represents the `device register personalized` command
var devicesRegisterPersonalizedCmd = &cobra.Command{
	Use:   "personalized [DevAddr] [NwkSKey] [AppSKey]",
	Short: "Create or update ABP registrations on the Handler",
	Long: `ttnctl devices register personalized creates or updates an ABP
registration on the Handler`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		appEUI := util.GetAppEUI(ctx)

		devAddr, err := types.ParseDevAddr(args[0])
		if err != nil {
			ctx.Fatalf("Invalid DevAddr: %s", err)
		}

		var nwkSKey types.NwkSKey
		var appSKey types.AppSKey
		if len(args) >= 3 {
			nwkSKey, err = types.ParseNwkSKey(args[1])
			if err != nil {
				ctx.Fatalf("Invalid NwkSKey: %s", err)
			}
			appSKey, err = types.ParseAppSKey(args[2])
			if err != nil {
				ctx.Fatalf("Invalid AppSKey: %s", err)
			}
			if reflect.DeepEqual(nwkSKey.Bytes(), defaultKey) || reflect.DeepEqual(appSKey.Bytes(), defaultKey) {
				ctx.Warn("You are using default keys, any attacker can read your data or attack your device's connectivity.")
			}
		} else {
			ctx.Info("Generating random NwkSKey and AppSKey...")
			copy(nwkSKey[:], random.Bytes(16))
			copy(appSKey[:], random.Bytes(16))
		}

		var flags uint32
		if value, _ := cmd.Flags().GetBool("relax-fcnt"); value {
			flags |= core.RelaxFcntCheck
			ctx.Warn("You are disabling frame counter checks. Your device is not protected against replay-attacks.")
		}

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found. Please login")
		}

		manager := util.GetHandlerManager(ctx)
		res, err := manager.UpsertABP(context.Background(), &core.UpsertABPHandlerReq{
			Token:   auth.AccessToken,
			AppEUI:  appEUI.Bytes(),
			DevAddr: devAddr.Bytes(),
			AppSKey: appSKey.Bytes(),
			NwkSKey: nwkSKey.Bytes(),
			Flags:   flags,
		})
		if err != nil || res == nil {
			ctx.WithError(err).Fatal("Could not register device")
		}
		ctx.WithFields(log.Fields{
			"DevAddr": devAddr,
			"NwkSKey": nwkSKey,
			"AppSKey": appSKey,
			"Flags":   flags,
		}).Info("Registered personalized device")
	},
}

// devicesRegisterDefaultCmd represents the `device register` command
var devicesRegisterDefaultCmd = &cobra.Command{
	Use:   "default [AppKey]",
	Short: "Create or update default OTAA registrations on the Handler",
	Long: `ttnctl devices register default creates or updates OTAA registrations
on the Handler that have not been explicitly registered using ttnctl devices
register [DevEUI] [AppKey]`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		appEUI := util.GetAppEUI(ctx)

		var appKey types.AppKey
		var err error
		if len(args) >= 2 {
			appKey, err = types.ParseAppKey(args[0])
			if err != nil {
				ctx.Fatalf("Invalid AppKey: %s", err)
			}
		} else {
			ctx.Info("Generating random AppKey...")
			copy(appKey[:], random.Bytes(16))
		}

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found. Please login")
		}

		manager := util.GetHandlerManager(ctx)
		res, err := manager.SetDefaultDevice(context.Background(), &core.SetDefaultDeviceReq{
			Token:  auth.AccessToken,
			AppEUI: appEUI.Bytes(),
			AppKey: appKey.Bytes(),
		})
		if err != nil || res == nil {
			ctx.WithError(err).Fatal("Could not set default device settings")
		}
		ctx.Info("Ok")
	},
}

func init() {
	RootCmd.AddCommand(devicesCmd)
	devicesCmd.AddCommand(devicesRegisterCmd)
	devicesCmd.AddCommand(devicesInfoCmd)
	devicesInfoCmd.Flags().Bool("lmic", false, "Print info for LMiC")
	devicesRegisterCmd.AddCommand(devicesRegisterPersonalizedCmd)
	devicesRegisterCmd.AddCommand(devicesRegisterDefaultCmd)
	devicesRegisterPersonalizedCmd.Flags().Bool("relax-fcnt", false, "Allow frame counter to reset (insecure)")
}
