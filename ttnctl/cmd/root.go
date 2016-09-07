// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/TheThingsNetwork/ttn/api"
	cliHandler "github.com/TheThingsNetwork/ttn/utils/cli/handler"
	"github.com/TheThingsNetwork/ttn/utils/logging"
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
		api.DialOptions = append(api.DialOptions, grpc.WithBlock())
		api.DialOptions = append(api.DialOptions, grpc.WithTimeout(2*time.Second))
		grpclog.SetLogger(logging.NewGRPCLogger(ctx))
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

	RootCmd.PersistentFlags().String("discovery-server", "discover.thethingsnetwork.org:1900", "The address of the Discovery server")
	viper.BindPFlag("discovery-server", RootCmd.PersistentFlags().Lookup("discovery-server"))

	RootCmd.PersistentFlags().String("ttn-router", "dev", "The ID of the TTN Router as announced in the Discovery server")
	viper.BindPFlag("ttn-router", RootCmd.PersistentFlags().Lookup("ttn-router"))

	RootCmd.PersistentFlags().String("ttn-handler", "dev", "The ID of the TTN Handler as announced in the Discovery server")
	viper.BindPFlag("ttn-handler", RootCmd.PersistentFlags().Lookup("ttn-handler"))

	RootCmd.PersistentFlags().String("mqtt-broker", "staging.thethingsnetwork.org:1883", "The address of the MQTT broker")
	viper.BindPFlag("mqtt-broker", RootCmd.PersistentFlags().Lookup("mqtt-broker"))

	RootCmd.PersistentFlags().String("ttn-account-server", "https://account.thethingsnetwork.org", "The address of the OAuth 2.0 server")
	viper.BindPFlag("ttn-account-server", RootCmd.PersistentFlags().Lookup("ttn-account-server"))
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
	viper.SetConfigType("yaml")
	viper.SetConfigName(".ttnctl")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("ttnctl") // set environment prefix
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error when reading config file:", err)
	} else if err == nil && viper.GetBool("debug") {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
