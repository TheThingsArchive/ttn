// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/net/context"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applicationsPayloadFunctionsCmd represents the applicationsPayloadFunctions command
var applicationsPayloadFunctionsCmd = &cobra.Command{
	Use:   "pf",
	Short: "Show the payload functions",
	Long: `ttnctl applications pf shows the payload functions for decoding,
converting and validating binary payload.
`,
	Run: func(cmd *cobra.Command, args []string) {
		appEUI := util.GetAppEUI(ctx)

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found. Please login")
		}

		manager := util.GetHandlerManager(ctx)
		res, err := manager.GetPayloadFunctions(context.Background(), &core.GetPayloadFunctionsReq{
			Token:  auth.AccessToken,
			AppEUI: appEUI.Bytes(),
		})
		if err != nil {
			ctx.WithError(err).Fatal("Could not get payload functions")
		}

		ctx.Info("Has decoder function")
		fmt.Println(res.Decoder)

		if res.Converter != "" {
			ctx.Info("Has converter function")
			fmt.Println(res.Converter)
		}

		if res.Validator != "" {
			ctx.Info("Has validator function")
			fmt.Println(res.Validator)
		} else {
			ctx.Warn("Does not have validator function")
		}
	},
}

// applicationsSetPayloadFunctionsCmd represents the applicationsSetPayloadFunctions command
var applicationsSetPayloadFunctionsCmd = &cobra.Command{
	Use:   "set [decoder.js]",
	Short: "Set payload functions",
	Long: `ttnctl applications pf set sets the decoder, converter and validator
function by loading the specified files containing JavaScript code.
`,
	Run: func(cmd *cobra.Command, args []string) {
		appEUI := util.GetAppEUI(ctx)

		if len(args) == 0 {
			cmd.Help()
			return
		}

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found. Please login")
		}

		decoder, err := ioutil.ReadFile(args[0])
		if err != nil {
			ctx.WithError(err).Fatal("Read decoder file failed")
		}

		var converter []byte
		converterFile, err := cmd.Flags().GetString("converter")
		if converterFile != "" {
			converter, err = ioutil.ReadFile(converterFile)
			if err != nil {
				ctx.WithError(err).Fatal("Read converter file failed")
			}
		}

		var validator []byte
		validatorFile, err := cmd.Flags().GetString("validator")
		if validatorFile != "" {
			validator, err = ioutil.ReadFile(validatorFile)
			if err != nil {
				ctx.WithError(err).Fatal("Read validator file failed")
			}
		}

		manager := util.GetHandlerManager(ctx)
		_, err = manager.SetPayloadFunctions(context.Background(), &core.SetPayloadFunctionsReq{
			Token:     auth.AccessToken,
			AppEUI:    appEUI.Bytes(),
			Decoder:   string(decoder),
			Converter: string(converter),
			Validator: string(validator),
		})
		if err != nil {
			ctx.WithError(err).Fatal("Could not set payload functions")
		}
		ctx.Info("Successfully set payload functions")
	},
}

// applicationsTestPayloadFunctionsCmd represents the applicationsTestPayloadFunctions command
var applicationsTestPayloadFunctionsCmd = &cobra.Command{
	Use:   "test [payload]",
	Short: "Test the payload functions",
	Long: `ttnctl applications pf test sends the specified binary data to the
Handler and returns the fields and validation result.
`,
	Run: func(cmd *cobra.Command, args []string) {
		appEUI := util.GetAppEUI(ctx)

		if len(args) == 0 {
			cmd.Help()
			return
		}

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found. Please login")
		}

		payload, err := util.ParseHEX(args[0], len(args[0]))
		if err != nil {
			ctx.WithError(err).Fatal("Invalid payload")
		}

		manager := util.GetHandlerManager(ctx)
		res, err := manager.TestPayloadFunctions(context.Background(), &core.TestPayloadFunctionsReq{
			Token:   auth.AccessToken,
			AppEUI:  appEUI.Bytes(),
			Payload: payload,
		})
		if err != nil {
			ctx.WithError(err).Fatal("Test payload functions failed")
		}

		if res.Valid {
			ctx.Info("Valid payload")
		} else {
			ctx.Warn("Invalid payload")
		}
		fmt.Printf("JSON: %s\n", res.Fields)
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsPayloadFunctionsCmd)
	applicationsPayloadFunctionsCmd.AddCommand(applicationsSetPayloadFunctionsCmd)
	applicationsPayloadFunctionsCmd.AddCommand(applicationsTestPayloadFunctionsCmd)

	applicationsSetPayloadFunctionsCmd.Flags().StringP("converter", "c", "", "Converter function")
	applicationsSetPayloadFunctionsCmd.Flags().StringP("validator", "v", "", "Validator function")
}
