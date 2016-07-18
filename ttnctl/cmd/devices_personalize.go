package cmd

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// devicesPersonalizeCmd represents the `device personalize` command
var devicesPersonalizeCmd = &cobra.Command{
	Use:   "personalize [Device ID] [DevAddr] [NwkSKey] [AppSKey]",
	Short: "Personalize a device",
	Long:  `ttnctl devices personalize can be used to personalize a device (ABP).`,
	Run: func(cmd *cobra.Command, args []string) {

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found, please login")
		}

		if len(args) == 0 {
			cmd.Usage()
			return
		}

		devID := args[0]
		if !api.ValidID(devID) {
			ctx.Fatalf("Invalid Device ID") // TODO: Add link to wiki explaining device IDs
		}

		appID := viper.GetString("app-id")
		if appID == "" {
			ctx.Fatal("Missing AppID. You should run ttnctl applications use [AppID] [AppEUI]")
		}

		var devAddr types.DevAddr
		if len(args) > 1 {
			devAddr, err = types.ParseDevAddr(args[1])
			if err != nil {
				ctx.Fatalf("Invalid DevAddr: %s", err)
			}
		} else {
			ctx.Info("Generating random DevAddr...")
			copy(devAddr[:], random.Bytes(8))
			devAddr[0] = (0x13 << 1) | (devAddr[0] & 1) // Use the TTN netID
		}

		var nwkSKey types.NwkSKey
		if len(args) > 2 {
			nwkSKey, err = types.ParseNwkSKey(args[2])
			if err != nil {
				ctx.Fatalf("Invalid NwkSKey: %s", err)
			}
		} else {
			ctx.Info("Generating random NwkSKey...")
			copy(nwkSKey[:], random.Bytes(16))
		}

		var appSKey types.AppSKey
		if len(args) > 3 {
			appSKey, err = types.ParseAppSKey(args[3])
			if err != nil {
				ctx.Fatalf("Invalid AppSKey: %s", err)
			}
		} else {
			ctx.Info("Generating random AppSKey...")
			copy(appSKey[:], random.Bytes(16))
		}

		manager, err := handler.NewManagerClient(viper.GetString("ttn-handler"), auth.AccessToken)
		if err != nil {
			ctx.WithError(err).Fatal("Could not create Handler client")
		}

		dev, err := manager.GetDevice(appID, devID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get existing device.")
		}

		dev.GetLorawanDevice().DevAddr = &devAddr
		dev.GetLorawanDevice().NwkSKey = &nwkSKey
		dev.GetLorawanDevice().AppSKey = &appSKey
		dev.GetLorawanDevice().FCntUp = 0
		dev.GetLorawanDevice().FCntDown = 0

		err = manager.SetDevice(dev)
		if err != nil {
			ctx.WithError(err).Fatal("Could not update Device")
		}

		ctx.WithFields(log.Fields{
			"AppID":   appID,
			"DevID":   devID,
			"DevAddr": devAddr,
			"NwkSKey": nwkSKey,
			"AppSKey": appSKey,
		}).Info("Personalized device")
	},
}

func init() {
	devicesCmd.AddCommand(devicesPersonalizeCmd)
}
