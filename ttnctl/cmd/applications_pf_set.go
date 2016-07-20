package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationsPayloadFunctionsSetCmd represents the `applications pf set` command
var applicationsPayloadFunctionsSetCmd = &cobra.Command{
	Use:   "pf set [decoder/converter/validator] [file.js]",
	Short: "Set payload functions of an application",
	Long: `ttnctl pf set can be used to get or set payload functions of an application.
The functions are read from the supplied file or from STDIN.`,
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

		app, err := manager.GetApplication(appID)
		if err != nil && strings.Contains(err.Error(), "not found") {
			app = &handler.Application{AppId: appID}
		} else if err != nil {
			ctx.WithError(err).Fatal("Could not get existing application.")
		}

		if len(args) == 0 {
			cmd.UsageFunc()(cmd)
			return
		}

		function := args[0]

		if len(args) == 2 {
			content, err := ioutil.ReadFile(args[1])
			if err != nil {
				ctx.WithError(err).Fatal("Could not read function file")
			}
			switch function {
			case "decoder":
				app.Decoder = string(content)
			case "converter":
				app.Converter = string(content)
			case "validator":
				app.Validator = string(content)
			default:
				ctx.Fatalf("Function %s does not exist", function)
			}
		} else {
			switch function {
			case "decoder":
				fmt.Println(`function (bytes) {
  // bytes is of type Buffer.

  // todo: return an object
  return {
    payload: bytes,
  };
}
########## Write your Decoder here and end with Ctrl+D (EOF):`)
				app.Decoder = readFunction()
			case "converter":
				fmt.Println(`function (val) {
  // val is the output of the decoder function.

  // todo: return an object
  return val;
}
########## Write your Converter here and end with Ctrl+D (EOF):`)
				app.Converter = readFunction()
			case "validator":
				fmt.Println(`function (val) {
  // val is the output of the converter function.

  // todo: return a boolean
  return true;
}
########## Write your Validator here and end with Ctrl+D (EOF):`)
				app.Validator = readFunction()
			default:
				ctx.Fatalf("Function %s does not exist", function)
			}
		}

		err = manager.SetApplication(app)
		if err != nil {
			ctx.WithError(err).Fatal("Could not update application")
		}

		ctx.WithFields(log.Fields{
			"AppID": appID,
		}).Infof("Updated application")
	},
}

func readFunction() string {
	content, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		ctx.WithError(err).Fatal("Could not read function from STDIN.")
	}
	return strings.TrimSpace(string(content))
}

func init() {
	applicationsCmd.AddCommand(applicationsPayloadFunctionsSetCmd)
}
