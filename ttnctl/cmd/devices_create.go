package cmd

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// devicesCreateCmd represents the `device create` command
var devicesCreateCmd = &cobra.Command{
	Use:   "create [Device ID] [DevEUI] [AppKey]",
	Short: "Create a new device",
	Long:  `ttnctl devices create can be used to create a new device.`,
	Run: func(cmd *cobra.Command, args []string) {

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found, please login")
		}

		if len(args) == 0 {
			ctx.Fatalf("Device ID is required")
		}

		devID := args[0]
		if !api.ValidID(devID) {
			ctx.Fatalf("Invalid Device ID") // TODO: Add link to wiki explaining device IDs
		}

		appID := viper.GetString("app-id")
		if appID == "" {
			ctx.Fatal("Missing AppID. You should run ttnctl applications use [AppID] [AppEUI]")
		}

		var appEUI types.AppEUI
		if suppliedAppEUI := viper.GetString("app-eui"); suppliedAppEUI != "" {
			appEUI, err = types.ParseAppEUI(suppliedAppEUI)
			if err != nil {
				ctx.Fatalf("Invalid AppEUI: %s", err)
			}
		} else {
			ctx.Fatal("Missing AppEUI. You should run ttnctl applications use [AppID] [AppEUI]")
		}

		var devEUI types.DevEUI
		if len(args) > 1 {
			devEUI, err = types.ParseDevEUI(args[1])
			if err != nil {
				ctx.Fatalf("Invalid DevEUI: %s", err)
			}
		} else {
			ctx.Info("Generating random DevEUI...")
			copy(devEUI[1:], random.Bytes(7))
		}

		var appKey types.AppKey
		if len(args) > 2 {
			appKey, err = types.ParseAppKey(args[2])
			if err != nil {
				ctx.Fatalf("Invalid AppKey: %s", err)
			}
		} else {
			ctx.Info("Generating random AppKey...")
			copy(appKey[:], random.Bytes(16))
		}

		manager, err := handler.NewManagerClient(viper.GetString("ttn-handler"), auth.AccessToken)
		if err != nil {
			ctx.WithError(err).Fatal("Could not create Handler client")
		}

		err = manager.SetDevice(&handler.Device{
			AppId: appID,
			DevId: devID,
			Device: &handler.Device_LorawanDevice{LorawanDevice: &lorawan.Device{
				AppId:  appID,
				DevId:  devID,
				AppEui: &appEUI,
				DevEui: &devEUI,
				AppKey: &appKey,
			}},
		})
		if err != nil {
			ctx.WithError(err).Fatal("Could not create Device")
		}

		ctx.WithFields(log.Fields{
			"AppID":  appID,
			"DevID":  devID,
			"AppEUI": appEUI,
			"DevEUI": devEUI,
			"AppKey": appKey,
		}).Info("Created device")
	},
}

func init() {
	devicesCmd.AddCommand(devicesCreateCmd)
}
