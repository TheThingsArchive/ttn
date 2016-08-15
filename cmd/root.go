// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/redis.v3"

	cliHandler "github.com/TheThingsNetwork/ttn/utils/cli/handler"
	"github.com/apex/log"
	jsonHandler "github.com/apex/log/handlers/json"
	levelHandler "github.com/apex/log/handlers/level"
	multiHandler "github.com/apex/log/handlers/multi"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var logFile *os.File

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

		var logHandlers []log.Handler
		logHandlers = append(logHandlers, levelHandler.New(cliHandler.New(os.Stdout), logLevel))

		if logFileLocation := viper.GetString("log-file"); logFileLocation != "" {
			absLogFileLocation, err := filepath.Abs(logFileLocation)
			if err != nil {
				panic(err)
			}
			logFile, err = os.OpenFile(absLogFileLocation, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				panic(err)
			}
			if err == nil {
				logHandlers = append(logHandlers, levelHandler.New(jsonHandler.New(logFile), logLevel))
			}
		}

		ctx = &log.Logger{
			Handler: multiHandler.New(logHandlers...),
		}
		ctx.WithFields(log.Fields{
			"ComponentID":     viper.GetString("id"),
			"Description":     viper.GetString("description"),
			"DiscoveryServer": viper.GetString("discovery-server"),
			"AuthServer":      viper.GetString("auth-server"),
		}).Info("Initializing The Things Network")
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if logFile != nil {
			logFile.Close()
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer func() {
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, false)
		if thePanic := recover(); thePanic != nil && ctx != nil {
			ctx.WithField("panic", thePanic).WithField("stack", string(buf)).Fatal("Stopping because of panic")
		}
	}()

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

	RootCmd.PersistentFlags().String("description", "", "The description of this component")
	viper.BindPFlag("description", RootCmd.PersistentFlags().Lookup("description"))

	RootCmd.PersistentFlags().String("discovery-server", "discover.thethingsnetwork.org:1900", "The address of the Discovery server")
	viper.BindPFlag("discovery-server", RootCmd.PersistentFlags().Lookup("discovery-server"))

	RootCmd.PersistentFlags().String("auth-server", "https://account.thethingsnetwork.org", "The address of the OAuth 2.0 server")
	viper.BindPFlag("auth-server", RootCmd.PersistentFlags().Lookup("auth-server"))

	RootCmd.PersistentFlags().String("auth-token", "", "The auth token signed JWT from the auth-server")
	viper.BindPFlag("auth-token", RootCmd.PersistentFlags().Lookup("auth-token"))

	RootCmd.PersistentFlags().Int("health-port", 0, "The port number where the health server should be started")
	viper.BindPFlag("health-port", RootCmd.PersistentFlags().Lookup("health-port"))

	dir, err := homedir.Dir()
	if err == nil {
		dir, _ = homedir.Expand(dir)
	}
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	RootCmd.PersistentFlags().Bool("tls", false, "Use TLS")
	viper.BindPFlag("tls", RootCmd.PersistentFlags().Lookup("tls"))

	RootCmd.PersistentFlags().String("key-dir", path.Clean(dir+"/.ttn/"), "The directory where public/private keys are stored")
	viper.BindPFlag("key-dir", RootCmd.PersistentFlags().Lookup("key-dir"))

	RootCmd.PersistentFlags().String("log-file", "", "Location of the log file")
	viper.BindPFlag("log-file", RootCmd.PersistentFlags().Lookup("log-file"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigType("yaml")
	viper.SetConfigName(".ttn")  // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.SetEnvPrefix("ttn")    // set environment prefix
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.BindEnv("debug")

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error when reading config file:", err)
	} else if err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

var RedisConnectRetries = 10
var RedisConnectRetryDelay = 1 * time.Second

func connectRedis(client *redis.Client) error {
	var err error
	for retries := 0; retries < RedisConnectRetries; retries++ {
		_, err = client.Ping().Result()
		if err == nil {
			break
		}
		ctx.WithError(err).Warn("Could not connect to Redis. Retrying...")
		<-time.After(RedisConnectRetryDelay)
	}
	if err != nil {
		client.Close()
		return err
	}
	return nil
}
