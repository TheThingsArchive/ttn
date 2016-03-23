// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/components/handler"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

const emptyCell = "-"

func getHandlerManager() core.AuthHandlerClient {
	cli, err := handler.NewClient(viper.GetString("ttn-handler"))
	if err != nil {
		ctx.Fatalf("Could not connect: %v", err)
	}
	return cli
}

// devicesCmd represents the `devices` command
var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage devices on the Handler",
	Long:  `ttnctl devices retrieves a list of devices that your application registered on the Handler.`,
	Run: func(cmd *cobra.Command, args []string) {
		appEUI, err := util.Parse64(viper.GetString("app-eui"))
		if err != nil {
			ctx.Fatalf("Invalid AppEUI: %s", err)
		}

		manager := getHandlerManager()

		res, err := manager.ListDevices(context.Background(), &core.ListDevicesHandlerReq{
			Token:  viper.GetString("app-token"),
			AppEUI: appEUI,
		})
		if err != nil {
			ctx.WithError(err).Fatal("Could not get device list")
		}

		ctx.Infof("Found %d personalized devices (ABP)", len(res.ABP))
		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("DevAddr", "NwkSKey", "AppSKey")
		for _, device := range res.ABP {
			devAddr := fmt.Sprintf("%X", device.DevAddr)
			nwkSKey := fmt.Sprintf("%X", device.NwkSKey)
			appSKey := fmt.Sprintf("%X", device.AppSKey)
			table.AddRow(devAddr, nwkSKey, appSKey)
		}
		fmt.Println(table)

		ctx.Infof("Found %d dynamic devices (OTAA)", len(res.OTAA))
		table = uitable.New()
		table.MaxColWidth = 40
		table.AddRow("DevEUI", "DevAddr", "NwkSKey", "AppSKey", "AppKey")
		for _, device := range res.OTAA {
			devEUI := fmt.Sprintf("%X", device.DevEUI)
			devAddr := fmt.Sprintf("%X", device.DevAddr)
			nwkSKey := fmt.Sprintf("%X", device.NwkSKey)
			appSKey := fmt.Sprintf("%X", device.AppSKey)
			appKey := fmt.Sprintf("%X", device.AppKey)
			table.AddRow(devEUI, devAddr, nwkSKey, appSKey, appKey)
		}
		fmt.Println(table)
	},
}

// devicesRegisterCmd represents the `device register` command
var devicesRegisterCmd = &cobra.Command{
	Use:   "register [DevEUI] [AppKey]",
	Short: "Create or Update registrations on the Handler",
	Long:  `ttnctl device register creates or updates an OTAA registration on the Handler`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			ctx.Fatal("Insufficient arguments")
		}

		appEUI, err := util.Parse64(viper.GetString("app-eui"))
		if err != nil {
			ctx.Fatalf("Invalid AppEUI: %s", err)
		}

		devEUI, err := util.Parse64(args[0])
		if err != nil {
			ctx.Fatalf("Invalid DevEUI: %s", err)
		}

		appKey, err := util.Parse128(args[1])
		if err != nil {
			ctx.Fatalf("Invalid AppKey: %s", err)
		}

		manager := getHandlerManager()
		res, err := manager.UpsertOTAA(context.Background(), &core.UpsertOTAAHandlerReq{
			Token:  viper.GetString("app-token"),
			AppEUI: appEUI,
			DevEUI: devEUI,
			AppKey: appKey,
		})
		if err != nil || res == nil {
			ctx.WithError(err).Fatal("Could not register device")
		}
		ctx.Info("Ok")
	},
}

// devicesRegisterPersonalizedCmd represents the `device register personalized` command
var devicesRegisterPersonalizedCmd = &cobra.Command{
	Use:   "personalized [DevAddr] [NwkSKey] [AppSKey]",
	Short: "Create or Update ABP registrations on the Handler",
	Long:  `ttnctl device register creates or updates an ABP registration on the Handler`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			ctx.Fatal("Insufficient arguments")
		}

		appEUI, err := util.Parse64(viper.GetString("app-eui"))
		if err != nil {
			ctx.Fatalf("Invalid AppEUI: %s", err)
		}

		devAddr, err := util.Parse32(args[0])
		if err != nil {
			ctx.Fatalf("Invalid DevAddr: %s", err)
		}

		nwkSKey, err := util.Parse128(args[1])
		if err != nil {
			ctx.Fatalf("Invalid NwkSKey: %s", err)
		}

		appSKey, err := util.Parse128(args[2])
		if err != nil {
			ctx.Fatalf("Invalid AppSKey: %s", err)
		}

		manager := getHandlerManager()
		res, err := manager.UpsertABP(context.Background(), &core.UpsertABPHandlerReq{
			Token:   viper.GetString("app-token"),
			AppEUI:  appEUI,
			DevAddr: devAddr,
			AppSKey: appSKey,
			NwkSKey: nwkSKey,
		})
		if err != nil || res == nil {
			ctx.WithError(err).Fatal("Could not register device")
		}
		ctx.Info("Ok")
	},
}

func init() {
	RootCmd.AddCommand(devicesCmd)

	devicesCmd.AddCommand(devicesRegisterCmd)

	devicesRegisterCmd.AddCommand(devicesRegisterPersonalizedCmd)

	devicesCmd.Flags().String("ttn-handler", "0.0.0.0:1782", "The net address of the TTN Handler")
	viper.BindPFlag("ttn-handler", devicesCmd.Flags().Lookup("ttn-handler"))

	devicesCmd.PersistentFlags().String("app-eui", "0102030405060708", "The app EUI to use")
	viper.BindPFlag("app-eui", devicesCmd.PersistentFlags().Lookup("app-eui"))

	devicesCmd.PersistentFlags().String("app-token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJUVE4tSEFORExFUi0xIiwiaXNzIjoiVGhlVGhpbmdzVGhlTmV0d29yayIsInN1YiI6IjAxMDIwMzA0MDUwNjA3MDgifQ.zMHNXAVgQj672lwwDVmfYshpMvPwm6A8oNWJ7teGS2A", "The app Token to use")
	viper.BindPFlag("app-token", devicesCmd.PersistentFlags().Lookup("app-token"))
}
