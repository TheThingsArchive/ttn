package cmd

import (
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationsRegisterCmd represents the `applications register` command
var applicationsRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register this application with the handler",
	Long:  `ttnctl register can be used to register this application with the handler.`,
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

		err = manager.RegisterApplication(appID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not register application")
		}

		ctx.WithFields(log.Fields{
			"AppID": appID,
		}).Infof("Registered application")
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsRegisterCmd)
}
