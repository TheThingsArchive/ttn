// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

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
)

var applicationsPayloadFunctionsSetCmd = &cobra.Command{
	Use:   "set [decoder/converter/validator/encoder] [file.js]",
	Short: "Set payload functions of an application",
	Long: `ttnctl pf set can be used to get or set payload functions of an application.
The functions are read from the supplied file or from STDIN.`,
	Run: func(cmd *cobra.Command, args []string) {

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx)
		defer conn.Close()

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
			case "encoder":
				app.Encoder = string(content)
			default:
				ctx.Fatalf("Function %s does not exist", function)
			}
		} else {
			switch function {
			case "decoder":
				fmt.Println(`function Decoder(bytes) {
  // Decode an uplink message from
  // a Buffer of bytes to an object.

  return {
    isLightOn: bytes[0]
  };
}
########## Write your Decoder here and end with Ctrl+D (EOF):`)
				app.Decoder = readFunction()
			case "converter":
				fmt.Println(`function Converter(decodedObj) {
  // Modify the decoded uplink message.
  
  decodedObj.isLightOn = !!decodedObj.isLightOn; 

  return decodedObj;
}
########## Write your Converter here and end with Ctrl+D (EOF):`)
				app.Converter = readFunction()
			case "validator":
				fmt.Println(`function Validator(convertedObj) {
  // Return false if the decoded and converted uplink
  // message is invalid and should be dropped.

  return true;
}
########## Write your Validator here and end with Ctrl+D (EOF):`)
				app.Validator = readFunction()
			case "encoder":
				fmt.Println(`function Encoder(obj) {
  // Convert uplink messages sent as object to
  // an array of bytes.
  
  return [ obj.turnLightOn ? 1 : 0 ];
}
########## Write your Encoder here and end with Ctrl+D (EOF):`)
				app.Encoder = readFunction()
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
	applicationsPayloadFunctionsCmd.AddCommand(applicationsPayloadFunctionsSetCmd)
}
