// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const emptyCell = "-"

func getHandlerManager() (*grpc.ClientConn, core.HandlerManagerClient) {
	conn, err := grpc.Dial(viper.GetString("ttn-handler"), grpc.WithInsecure())
	if err != nil {
		ctx.Fatalf("Could not connect: %v", err)
	}

	return conn, core.NewHandlerManagerClient(conn)
}

// devicesCmd represents the `devices` command
var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List registered devices on the Handler",
	Long:  `List registered devices on the Handler`,
	Run: func(cmd *cobra.Command, args []string) {
		appEUI, err := util.Parse64(viper.GetString("app-eui"))
		if err != nil {
			ctx.Fatalf("Invalid AppEUI: %s", err)
		}

		conn, manager := getHandlerManager()
		defer conn.Close()

		res, err := manager.ListDevices(context.Background(), &core.ListDevicesHandlerReq{
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
	Short: "Create or Update OTAA registrations on the Handler",
	Long:  `Create or Update OTAA registrations on the Handler`,
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

		conn, manager := getHandlerManager()
		defer conn.Close()

		res, err := manager.UpsertOTAA(context.Background(), &core.UpsertOTAAHandlerReq{
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
	Long:  `Create or Update ABP registrations on the Handler`,
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

		conn, manager := getHandlerManager()
		defer conn.Close()

		res, err := manager.UpsertABP(context.Background(), &core.UpsertABPHandlerReq{
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

	devicesCmd.Flags().String("ttn-handler", "0.0.0.0:1882", "The net address of the TTN Handler")
	viper.BindPFlag("ttn-handler", devicesCmd.Flags().Lookup("ttn-handler"))

	devicesCmd.PersistentFlags().String("app-eui", "0102030405060708", "The app EUI to use")
	viper.BindPFlag("app-eui", devicesCmd.PersistentFlags().Lookup("app-eui"))

	devicesCmd.PersistentFlags().String("app-token", "0102030405060708", "The app Token to use")
	viper.BindPFlag("app-token", devicesCmd.PersistentFlags().Lookup("app-token"))
}
