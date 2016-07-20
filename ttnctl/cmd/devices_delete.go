package cmd

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// devicesDeleteCmd represents the `device delete` command
var devicesDeleteCmd = &cobra.Command{
	Use:   "delete [Device ID]",
	Short: "Delete a device",
	Long:  `ttnctl devices delete can be used to delete a device.`,
	Run: func(cmd *cobra.Command, args []string) {

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found, please login")
		}

		if len(args) == 0 {
			cmd.UsageFunc()(cmd)
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

		manager, err := handler.NewManagerClient(viper.GetString("ttn-handler"), auth.AccessToken)
		if err != nil {
			ctx.WithError(err).Fatal("Could not create Handler client")
		}

		err = manager.DeleteDevice(appID, devID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not delete device.")
		}

		ctx.WithFields(log.Fields{
			"AppID": appID,
			"DevID": devID,
		}).Info("Deleted device")
	},
}

func init() {
	devicesCmd.AddCommand(devicesDeleteCmd)
}
