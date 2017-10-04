// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/api"
	"github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/go-account-lib/rights"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/random"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

const registerOnJoinAccessKeyName = "register-on-join"

var devicesRegisterOnJoinCmd = &cobra.Command{
	Use:   "on-join [Device ID Prefix] [AppKey]",
	Short: "Register a new device on join",
	Long:  `ttnctl devices register on-join can be used to register a device template for on-join registrations.`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 1, 2)

		var err error

		devID := strings.ToLower(args[0])
		if err := api.NotEmptyAndValidID(devID, "Device ID Prefix"); err != nil {
			ctx.Fatal(err.Error())
		}
		if len(devID) > 19 { // IDs will be (devID-eui) -> 36 - 16 - 1
			ctx.Fatal("Device ID Prefix is too long for on-join registration. The maximum length is 19.")
		}

		appID := util.GetAppID(ctx)
		appEUI := util.GetAppEUI(ctx)

		var devEUI types.DevEUI

		var appKey types.AppKey
		if len(args) > 1 {
			appKey, err = types.ParseAppKey(args[1])
			if err != nil {
				ctx.Fatalf("Invalid AppKey: %s", err)
			}
		} else {
			ctx.Info("Generating random AppKey...")
			random.FillBytes(appKey[:])
		}

		ctx.Warn("On-join registration does not follow LoRaWAN best practices. Unless you implement additional security")
		ctx.Warn("measures, it may corrupt your entire application if even one of your devices gets compromised.")

		device := &handler.Device{
			AppID:       appID,
			DevID:       devID,
			Description: fmt.Sprintf("Device template for on-join registrations for %s", appEUI),
			Device: &handler.Device_LoRaWANDevice{LoRaWANDevice: &lorawan.Device{
				AppID:         appID,
				DevID:         devID,
				AppEUI:        appEUI,
				DevEUI:        devEUI,
				AppKey:        &appKey,
				Uses32BitFCnt: true,
			}},
		}

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		app, err := manager.GetApplication(appID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get Application from Handler")
		}
		if app.RegisterOnJoinAccessKey == "" {
			ctx.Info("Application does not have a RegisterOnJoinAccessKey")
			ctx.Info("Looking for existing key...")

			account := util.GetAccount(ctx)
			accountApp, err := account.FindApplication(appID)
			if err != nil {
				ctx.WithError(err).Fatal("Could not get Application from Account")
			}

			for _, key := range accountApp.AccessKeys {
				if key.Name == registerOnJoinAccessKeyName {
					ctx.Info("Using existing key...")
					app.RegisterOnJoinAccessKey = key.Key
					break
				}
			}

			if app.RegisterOnJoinAccessKey == "" {
				ctx.Info("No key found")
				ctx.Info("Requesting a new key...")
				key, err := account.AddAccessKey(appID, registerOnJoinAccessKeyName, []types.Right{rights.Devices})
				if err != nil {
					ctx.WithError(err).Fatal("Could not create RegisterOnJoinAccessKey")
				}
				app.RegisterOnJoinAccessKey = key.Key
				ctx.Info("New key registered")
			}

			err = manager.SetApplication(app)
			if err != nil {
				ctx.WithError(err).Fatal("Could not set Application on Handler")
			}
		}

		err = manager.SetDevice(device)
		if err != nil {
			ctx.WithError(err).Fatal("Could not register Device")
		}

		ctx.WithFields(ttnlog.Fields{
			"AppID":  appID,
			"DevID":  devID,
			"AppEUI": appEUI,
			"AppKey": appKey,
		}).Info("Registered on-join device template")
	},
}

func init() {
	devicesRegisterCmd.AddCommand(devicesRegisterOnJoinCmd)
}
