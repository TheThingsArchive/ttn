// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

var applicationsPayloadFunctionsSetCmd = &cobra.Command{
	Use:   "set [decoder/converter/validator/encoder] [file.js]",
	Short: "Set payload functions of an application",
	Long: `ttnctl pf set can be used to get or set payload functions of an application.
The functions are read from the supplied file or from STDIN.`,
	Example: `$ ttnctl applications pf set decoder
  INFO Discovering Handler...
  INFO Connecting with Handler...
function Decoder(bytes, port) {
  // Decode an uplink message from a buffer
  // (array) of bytes to an object of fields.
  var decoded = {};

  // if (port === 1) {
  //   decoded.led = bytes[0];
  // }

  return decoded;
}
########## Write your Decoder here and end with Ctrl+D (EOF):
function Decoder(bytes, port) {
  var decoded = {};

  // if (port === 1) {
  //   decoded.led = bytes[0];
  // }

  return decoded;
}

Do you want to test the payload functions? (Y/n)
Y

Payload: 12 34
Port: 1

  INFO Function tested successfully

  INFO Updated application                      AppID=test
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 2)

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		app, err := manager.GetApplication(appID)
		if err != nil && strings.Contains(err.Error(), "not found") {
			app = &handler.Application{AppId: appID}
		} else if err != nil {
			ctx.WithError(err).Fatal("Could not get existing application.")
		}

		function := args[0]

		if len(args) == 2 {
			content, err := ioutil.ReadFile(args[1])
			if err != nil {
				ctx.WithError(err).Fatal("Could not read function file")
			}
			fmt.Println(fmt.Sprintf(`
Function read from %s:

%s
`, args[1], string(content)))

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
				fmt.Println(`function Decoder(bytes, port) {
  // Decode an uplink message from a buffer
  // (array) of bytes to an object of fields.
  var decoded = {};

  // if (port === 1) {
  //   decoded.led = bytes[0];
  // }

  return decoded;
}
########## Write your Decoder here and end with Ctrl+D (EOF):`)
				app.Decoder = readFunction(ctx)
			case "converter":
				fmt.Println(`function Converter(decoded, port) {
  // Merge, split or otherwise
  // mutate decoded fields.
  var converted = decoded;

  // if (port === 1 && (converted.led === 0 || converted.led === 1)) {
  //   converted.led = Boolean(converted.led);
  // }

  return converted;
}
########## Write your Converter here and end with Ctrl+D (EOF):`)
				app.Converter = readFunction(ctx)
			case "validator":
				fmt.Println(`function Validator(converted, port) {
  // Return false if the decoded, converted
  // message is invalid and should be dropped.

  // if (port === 1 && typeof converted.led !== 'boolean') {
  //   return false;
  // }

  return true;
}
########## Write your Validator here and end with Ctrl+D (EOF):`)
				app.Validator = readFunction(ctx)
			case "encoder":
				fmt.Println(`function Encoder(object, port) {
  // Encode downlink messages sent as
  // object to an array or buffer of bytes.
  var bytes = [];

  // if (port === 1) {
  //   bytes[0] = object.led ? 1 : 0;
  // }

  return bytes;
}
########## Write your Encoder here and end with Ctrl+D (EOF):`)
				app.Encoder = readFunction(ctx)
			default:
				ctx.Fatalf("Function %s does not exist", function)
			}
		}

		fmt.Printf("\nDo you want to test the payload functions? (Y/n)\n")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" || response == "" {
			switch function {
			case "decoder", "converter", "validator":
				payload, err := util.ReadPayload()
				if err != nil {
					ctx.WithError(err).Fatal("Could not parse the payload")
				}

				port, err := util.ReadPort()
				if err != nil {
					ctx.WithError(err).Fatal("Could not parse the port")
				}

				result, err := manager.DryUplink(payload, app, uint32(port))
				if err != nil {
					ctx.WithError(err).Fatal("Could not set the payload function")
				}

				if !result.Valid {
					ctx.Fatal("Could not set the payload function: Invalid result")
				}
				ctx.Infof("Function tested successfully. Object returned by the converter: %s", result.Fields)
			case "encoder":
				fields, err := util.ReadFields()
				if err != nil {
					ctx.WithError(err).Fatal("Could not parse the fields")
				}

				port, err := util.ReadPort()
				if err != nil {
					ctx.WithError(err).Fatal("Could not parse the port")
				}

				result, err := manager.DryDownlinkWithFields(fields, app, uint32(port))
				if err != nil {
					ctx.WithError(err).Fatal("Could not set the payload function")
				}
				ctx.Infof("Function tested successfully. Encoded message: %v", result.Payload)
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

func init() {
	applicationsPayloadFunctionsCmd.AddCommand(applicationsPayloadFunctionsSetCmd)
}

func readFunction(ctx log.Interface) string {
	content, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		ctx.WithError(err).Fatal("Could not read function from STDIN.")
	}
	return strings.TrimSpace(string(content))
}
