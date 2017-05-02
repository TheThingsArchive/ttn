// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	cliHandler "github.com/TheThingsNetwork/go-utils/handlers/cli"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/go-utils/log/grpc"
	"github.com/TheThingsNetwork/ttn/api"
	esHandler "github.com/TheThingsNetwork/ttn/utils/elasticsearch/handler"
	"github.com/apex/log"
	jsonHandler "github.com/apex/log/handlers/json"
	levelHandler "github.com/apex/log/handlers/level"
	multiHandler "github.com/apex/log/handlers/multi"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tj/go-elastic"
	"google.golang.org/grpc/grpclog"
	"gopkg.in/redis.v5"
)

var cfgFile string

var logFile *os.File

var ctx ttnlog.Interface

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

		if !viper.GetBool("no-cli-logs") {
			logHandlers = append(logHandlers, levelHandler.New(cliHandler.New(os.Stdout), logLevel))
		}

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

		if esServer := viper.GetString("elasticsearch"); esServer != "" {
			esClient := elastic.New(esServer)
			esClient.HTTPClient = &http.Client{
				Timeout: 5 * time.Second,
			}
			logHandlers = append(logHandlers, levelHandler.New(esHandler.New(&esHandler.Config{
				Client:     esClient,
				Prefix:     cmd.Name(),
				BufferSize: 10,
			}), logLevel))
		}

		// Set the API/gRPC logger
		ctx = apex.Wrap(&log.Logger{
			Handler: multiHandler.New(logHandlers...),
		})
		ttnlog.Set(ctx)
		grpclog.SetLogger(grpc.Wrap(ttnlog.Get()))

		if viper.GetBool("allow-insecure") {
			api.AllowInsecureFallback = true
		}

		ctx.WithFields(ttnlog.Fields{
			"ComponentID":              viper.GetString("id"),
			"Description":              viper.GetString("description"),
			"Discovery Server Address": viper.GetString("discovery-address"),
			"Auth Servers":             viper.GetStringMapString("auth-servers"),
			"Monitors":                 viper.GetStringMapString("monitor-servers"),
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
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default \"$HOME/.ttn.yml\")")

	RootCmd.PersistentFlags().Bool("no-cli-logs", false, "Disable CLI logs")
	RootCmd.PersistentFlags().String("log-file", "", "Location of the log file")
	RootCmd.PersistentFlags().String("elasticsearch", "", "Location of Elasticsearch server for logging")

	RootCmd.PersistentFlags().String("id", "", "The id of this component")
	RootCmd.PersistentFlags().String("description", "", "The description of this component")
	RootCmd.PersistentFlags().Bool("public", false, "Announce this component as part of The Things Network (public community network)")

	RootCmd.PersistentFlags().String("discovery-address", "discover.thethingsnetwork.org:1900", "The address of the Discovery server")
	RootCmd.PersistentFlags().String("auth-token", "", "The JWT token to be used for the discovery server")

	RootCmd.PersistentFlags().Int("health-port", 0, "The port number where the health server should be started")

	viper.SetDefault("auth-servers", map[string]string{
		"ttn-account-v2": "https://account.thethingsnetwork.org",
	})

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

	RootCmd.PersistentFlags().Bool("tls", true, "Use TLS")
	RootCmd.PersistentFlags().Bool("allow-insecure", false, "Allow insecure fallback if TLS unavailable")
	RootCmd.PersistentFlags().String("key-dir", path.Clean(dir+"/.ttn/"), "The directory where public/private keys are stored")

	viper.BindPFlags(RootCmd.PersistentFlags())
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
		os.Exit(1)
	} else if err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// RedisConnectRetries indicates how many times the Redis connection should be retried
var RedisConnectRetries = 10

// RedisConnectRetryDelay indicates the time between Redis connection retries
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
