// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	cliHandler "github.com/TheThingsNetwork/ttn/utils/cli/handler"
	"github.com/apex/log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var ctx log.Interface

// RootCmd is executed when ttn is executed without a subcommand
var RootCmd = &cobra.Command{
	Use:   "ttn",
	Short: "The Things Network's backend servers",
	Long:  `ttn launches The Things Network's backend servers`,
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

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default \"$HOME/.ttn.yaml\")")

	RootCmd.PersistentFlags().String("id", "", "The id of this component")
	viper.BindPFlag("id", RootCmd.PersistentFlags().Lookup("id"))

	RootCmd.PersistentFlags().String("token", "", "The auth token this component should use")
	viper.BindPFlag("token", RootCmd.PersistentFlags().Lookup("token"))

	RootCmd.PersistentFlags().String("description", "", "The description of this component")
	viper.BindPFlag("description", RootCmd.PersistentFlags().Lookup("description"))

	RootCmd.PersistentFlags().String("discovery-server", "localhost:1900", "The address of the Discovery server")
	viper.BindPFlag("discovery-server", RootCmd.PersistentFlags().Lookup("discovery-server"))

	RootCmd.PersistentFlags().String("auth-server", "https://account.thethingsnetwork.org", "The address of the OAuth 2.0 server")
	viper.BindPFlag("auth-server", RootCmd.PersistentFlags().Lookup("auth-server"))

	var defaultOAuth2KeyFile string
	dir, err := homedir.Dir()
	if err == nil {
		expanded, err := homedir.Expand(dir)
		if err == nil {
			defaultOAuth2KeyFile = path.Join(expanded, ".ttn/oauth2-token.pub")
		}
	}

	RootCmd.PersistentFlags().String("oauth2-keyfile", defaultOAuth2KeyFile, "The OAuth 2.0 public key")
	viper.BindPFlag("oauth2-keyfile", RootCmd.PersistentFlags().Lookup("oauth2-keyfile"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".ttn")  // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.SetEnvPrefix("ttn")    // set environment prefix
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	viper.BindEnv("debug")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
