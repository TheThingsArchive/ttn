// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
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
	Parsing function...

	Test the function to detect runtime errors
	Provide the signature of the payload function with test values

	Note:
	1) Use single quotes for strings: E.g: 'this is a valid string'
	2) Use the built-in function JSON.stringify() to provide json objects parameters: E.g: JSON.stingify({ valid: argument })
	3) The provided signature should match the previous function declaration: E.g: MyFunc('entry', 123) will allow us to test the function called MyFunc() and which takes 2 arguments.
	########## Write your testing entry here and end with Ctrl+D (EOF):
	Decoder([10, 32], 3)
	  INFO Testing...

	The test is successful, the given function is a valid payload
	  INFO Updated application                      AppID=test
	`,
	Run: func(cmd *cobra.Command, args []string) {

		appID := util.GetAppID(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)
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
			fmt.Println(fmt.Sprintf(`
Function read from %s:

%s
`, args[1], string(content)))

			code, err := util.ValidatePayload(ctx, string(content), function)
			if err != nil {
				ctx.WithError(err).Fatal("Could not validate the function.")
			}
			switch function {
			case "decoder":
				app.Decoder = code
			case "converter":
				app.Decoder = code
			case "encoder":
				app.Encoder = code
			case "validator":
				app.Validator = code
			default:
				ctx.Fatalf("Function %s does not exist", function)
			}
		} else {
			switch function {
			case "decoder":
				fmt.Println(`
function Decoder(bytes, port) {
// Decode an uplink message from a buffer
// (array) of bytes to an object of fields.
var decoded = {};

// if (port === 1) {
//   decoded.led = bytes[0];
// }

return decoded;
}
########## Write your Decoder here and end with Ctrl+D (EOF):`)
				code, err := util.ValidatePayload(ctx, util.ReadFunction(ctx), function)
				if err != nil {
					ctx.WithError(err).Fatal("Could not validate the function")
				}
				app.Decoder = code
			case "converter":
				fmt.Println(`
function Converter(decoded, port) {
// Merge, split or otherwise
// mutate decoded fields.
var converted = decoded;

// if (port === 1 && (converted.led === 0 || converted.led === 1)) {
//   converted.led = Boolean(converted.led);
// }

return converted;
}
########## Write your Converter here and end with Ctrl+D (EOF):`)
				code, err := util.ValidatePayload(ctx, util.ReadFunction(ctx), function)
				if err != nil {
					ctx.WithError(err).Fatal("Could not validate the function")
				}
				app.Converter = code
			case "validator":
				fmt.Println(`
function Validator(converted, port) {
// Return false if the decoded, converted
// message is invalid and should be dropped.

// if (port === 1 && typeof converted.led !== 'boolean') {
//   return false;
// }

return true;
}
########## Write your Validator here and end with Ctrl+D (EOF):`)
				code, err := util.ValidatePayload(ctx, util.ReadFunction(ctx), function)
				if err != nil {
					ctx.WithError(err).Fatal("Could not validate the function")
				}
				app.Validator = code
			case "encoder":
				fmt.Println(`
function Encoder(object, port) {
// Encode downlink messages sent as
// object to an array or buffer of bytes.
var bytes = [];

// if (port === 1) {
//   bytes[0] = object.led ? 1 : 0;
// }

return bytes;
}
########## Write your Encoder here and end with Ctrl+D (EOF):`)
				code, err := util.ValidatePayload(ctx, util.ReadFunction(ctx), function)
				if err != nil {
					ctx.WithError(err).Fatal("Could not validate the function")
				}
				app.Encoder = code
			default:
				ctx.Fatalf("Function %s does not exist", function)
			}
		}

		err = manager.SetApplication(app)
		if err != nil {
			ctx.WithError(err).Fatal("Could not update application")
		}

		ctx.WithFields(ttnlog.Fields{
			"AppID": appID,
		}).Infof("Updated application")
	},
}

func init() {
	applicationsPayloadFunctionsCmd.AddCommand(applicationsPayloadFunctionsSetCmd)
}
