// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	cliHandler "github.com/TheThingsNetwork/go-utils/handlers/cli"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/TheThingsNetwork/ttn/utils/logging"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var cfgFile string
var dataDir string
var debug bool

var ctx log.Interface

// RootCmd is the entrypoint for ttnctl
var RootCmd = &cobra.Command{
	Use:   "ttnctl",
	Short: "Control The Things Network from the command line",
	Long:  `ttnctl controls The Things Network from the command line.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var logLevel = log.InfoLevel
		if debug {
			logLevel = log.DebugLevel
		}
		ctx = &log.Logger{
			Level:   logLevel,
			Handler: cliHandler.New(os.Stdout),
		}

		if debug {
			util.PrintConfig(ctx, true)
		}

		api.DialOptions = append(api.DialOptions, grpc.WithBlock())
		api.DialOptions = append(api.DialOptions, grpc.WithTimeout(2*time.Second))
		grpclog.SetLogger(logging.NewGRPCLogger(ctx))
		api.SetLogger(api.Apex(ctx))
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

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ttnctl.yml)")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	RootCmd.PersistentFlags().StringVar(&dataDir, "data", "", "directory where ttnctl stores data (default is $HOME/.ttnctl)")
	viper.BindPFlag("data", RootCmd.PersistentFlags().Lookup("data"))

	RootCmd.PersistentFlags().String("discovery-address", "discover.thethingsnetwork.org:1900", "The address of the Discovery server")
	viper.BindPFlag("discovery-address", RootCmd.PersistentFlags().Lookup("discovery-address"))

	RootCmd.PersistentFlags().String("router-id", "ttn-router-eu", "The ID of the TTN Router as announced in the Discovery server")
	viper.BindPFlag("router-id", RootCmd.PersistentFlags().Lookup("router-id"))

	RootCmd.PersistentFlags().String("handler-id", "ttn-handler-eu", "The ID of the TTN Handler as announced in the Discovery server")
	viper.BindPFlag("handler-id", RootCmd.PersistentFlags().Lookup("handler-id"))

	RootCmd.PersistentFlags().String("mqtt-address", "eu.thethings.network:1883", "The address of the MQTT broker")
	viper.BindPFlag("mqtt-address", RootCmd.PersistentFlags().Lookup("mqtt-address"))

	RootCmd.PersistentFlags().String("mqtt-username", "", "The username for the MQTT broker")
	viper.BindPFlag("mqtt-username", RootCmd.PersistentFlags().Lookup("mqtt-username"))

	RootCmd.PersistentFlags().String("mqtt-password", "", "The password for the MQTT broker")
	viper.BindPFlag("mqtt-password", RootCmd.PersistentFlags().Lookup("mqtt-password"))

	RootCmd.PersistentFlags().String("auth-server", "https://account.thethingsnetwork.org", "The address of the OAuth 2.0 server")
	viper.BindPFlag("auth-server", RootCmd.PersistentFlags().Lookup("auth-server"))
}

func printKV(key, t interface{}) {
	var val string
	switch t := t.(type) {
	case []byte:
		val = fmt.Sprintf("%X", t)
	default:
		val = fmt.Sprintf("%v", t)
	}

	if val != "" {
		fmt.Printf("%20s: %s\n", key, val)
	}
}

func confirm(prompt string) bool {
	fmt.Println(prompt)
	fmt.Print("> ")
	var answer string
	fmt.Scanf("%s", &answer)
	switch strings.ToLower(answer) {
	case "yes", "y":
		return true
	case "no", "n", "":
		return false
	default:
		fmt.Println("When you make up your mind, please answer yes or no.")
		return confirm(prompt)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "" {
		cfgFile = util.GetConfigFile()
	}

	if dataDir == "" {
		dataDir = util.GetDataDir()
	}

	viper.SetConfigType("yaml")
	viper.SetConfigFile(cfgFile)

	viper.SetEnvPrefix("ttnctl") // set environment prefix
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if _, err := os.Stat(cfgFile); err == nil {
		err := viper.ReadInConfig()
		if err != nil {
			fmt.Println("Error when reading config file:", err)
		}
	}
}
