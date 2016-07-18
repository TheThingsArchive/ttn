package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/apex/log"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// devicesListCmd represents the `device list` command
var devicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List al devices for the current application",
	Long:  `ttnctl devices list can be used to list all devices for the current application.`,
	Run: func(cmd *cobra.Command, args []string) {

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found, please login")
		}

		appID := viper.GetString("app-id")
		if appID == "" {
			ctx.Fatal("Missing AppID. You should run ttnctl applications use [AppID] [AppEUI]")
		}

		manager, err := handler.NewManagerClient(viper.GetString("ttn-handler"), auth.AccessToken)
		if err != nil {
			ctx.WithError(err).Fatal("Could not create Handler client")
		}

		devices, err := manager.GetDevicesForApplication(appID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get devices.")
		}

		table := uitable.New()
		table.MaxColWidth = 70
		table.AddRow("DevID", "AppEUI", "DevEUI", "DevAddr", "Up/Down")
		for _, dev := range devices {
			if lorawan := dev.GetLorawanDevice(); lorawan != nil {
				devAddr := lorawan.DevAddr
				if devAddr.IsEmpty() {
					devAddr = nil
				}
				table.AddRow(dev.DevId, lorawan.AppEui, lorawan.DevEui, devAddr, fmt.Sprintf("%d/%d", lorawan.FCntUp, lorawan.FCntDown))
			} else {
				table.AddRow(dev.DevId)
			}
		}

		fmt.Println()
		fmt.Println(table)
		fmt.Println()

		ctx.WithFields(log.Fields{
			"AppID": appID,
		}).Infof("Listed %d devices", len(devices))
	},
}

func init() {
	devicesCmd.AddCommand(devicesListCmd)
}
