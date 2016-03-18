// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/ttnctl/mqtt"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// registerCmd represents the `register` command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register an entity with TTN",
	Long:  `Register a user`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Fatal("register is not implemented.")
	},
}

// registerPersonalizedDeviceCmd represents the `register personalized-device` command
var registerPersonalizedDeviceCmd = &cobra.Command{
	Use:   "personalized-device [DevAddr] [NwkSKey] [AppSKey]",
	Short: "Register a new personalized device",
	Long:  `Register a new personalized device`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			ctx.Fatal("Insufficient arguments")
		}

		devAddr, err := util.Parse32(args[0])
		if err != nil {
			ctx.Fatalf("Invalid DevAddr: %s", err)
		}

		nwkSKey, err := util.Parse128(args[1])
		if err != nil {
			ctx.Fatalf("Invalid NwkSKey: %s", err)
		}

		appSKey, err := util.Parse128(args[2])
		if err != nil {
			ctx.Fatalf("Invalid AppSKey: %s", err)
		}

		payload, err := core.ABPSubAppReq{
			DevAddr: args[0],
			NwkSKey: args[1],
			AppSKey: args[2],
		}.MarshalMsg(nil)

		if err != nil {
			ctx.WithError(err).Fatal("Unable to create a registration")
		}

		mqtt.Setup(viper.GetString("handler.mqtt-broker"), ctx)
		mqtt.Connect()

		ctx.WithFields(log.Fields{
			"DevAddr": hex.EncodeToString(devAddr),
			"NwkSKey": hex.EncodeToString(nwkSKey),
			"AppSKey": hex.EncodeToString(appSKey),
		}).Info("Registering device...")

		token := mqtt.Client.Publish(fmt.Sprintf("%s/devices/personalized/activations", viper.GetString("handler.app-eui")), 2, false, payload)
		if token.Wait() && token.Error() != nil {
			ctx.WithError(token.Error()).Fatal("Registration failed.")
		} else {
			// Although we can't be sure whether it actually succeeded, we can know when the command is published to the MQTT.
			ctx.Info("Registration finished.")
		}

	},
}

func init() {
	RootCmd.AddCommand(registerCmd)

	registerCmd.AddCommand(registerPersonalizedDeviceCmd)

	registerPersonalizedDeviceCmd.Flags().String("mqtt-broker", "localhost:1883", "The address of the MQTT broker")
	viper.BindPFlag("handler.mqtt-broker", registerPersonalizedDeviceCmd.Flags().Lookup("mqtt-broker"))

	registerPersonalizedDeviceCmd.Flags().String("app-eui", "0102030405060708", "The app EUI to use")
	viper.BindPFlag("handler.app-eui", registerPersonalizedDeviceCmd.Flags().Lookup("app-eui"))

}
