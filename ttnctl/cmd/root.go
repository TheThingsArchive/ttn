// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"strings"

	cliHandler "github.com/TheThingsNetwork/ttn/utils/cli/handler"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var ctx log.Interface

// RootCmd is the entrypoint for handlerctl
var RootCmd = &cobra.Command{
	Use:   "ttnctl",
	Short: "Control The Things Network from the command line",
	Long:  `ttnctl controls The Things Network from the command line.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var logLevel = log.InfoLevel
		if viper.GetBool("debug") {
			logLevel = log.DebugLevel
		}
		ctx = &log.Logger{
			Level:   logLevel,
			Handler: cliHandler.New(os.Stdout),
		}
	},
}

// Execute runs on start
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// init initializes the configuration and command line flags
func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ttnctl.yaml)")

	RootCmd.PersistentFlags().String("ttn-router", "staging.thethingsnetwork.org:1700", "The net address of the TTN Router")
	viper.BindPFlag("ttn-router", RootCmd.PersistentFlags().Lookup("ttn-router"))

	RootCmd.PersistentFlags().String("ttn-handler", "staging.thethingsnetwork.org:1782", "The net address of the TTN Handler")
	viper.BindPFlag("ttn-handler", RootCmd.PersistentFlags().Lookup("ttn-handler"))

	RootCmd.PersistentFlags().String("mqtt-broker", "staging.thethingsnetwork.org:1883", "The address of the MQTT broker")
	viper.BindPFlag("mqtt-broker", RootCmd.PersistentFlags().Lookup("mqtt-broker"))

	RootCmd.PersistentFlags().String("app-eui", "", "The app EUI to use")
	viper.BindPFlag("app-eui", RootCmd.PersistentFlags().Lookup("app-eui"))

	RootCmd.PersistentFlags().String("ttn-account-server", "https://account.thethingsnetwork.org", "The address of the OAuth 2.0 server")
	viper.BindPFlag("ttn-account-server", RootCmd.PersistentFlags().Lookup("ttn-account-server"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".ttnctl")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("ttnctl") // set environment prefix
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && viper.GetBool("debug") {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
