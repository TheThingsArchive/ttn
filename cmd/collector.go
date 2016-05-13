// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/TheThingsNetwork/ttn/core/components/collector"
	"github.com/TheThingsNetwork/ttn/core/components/collector/influxdb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// collectorCmd represents the collector command
var collectorCmd = &cobra.Command{
	Use:   "collector",
	Short: "The Things Network collector",
	Long: `ttn collector starts the Collector component of The Things Network.

The Collector is responsible for storing uplink packets from the handler for
configured applications.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Info("Starting")

		appStorage, err := collector.ConnectRedis(viper.GetString("collector.redis-addr"), int64(viper.GetInt("collector.redis-db")))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to connect to Redis")
		}
		defer appStorage.Close()

		dataStorage, err := influxdb.NewDataStorage(viper.GetString("collector.influxdb-addr"),
			viper.GetString("collector.influxdb-username"), viper.GetString("collector.influxdb-password"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to connect to InfluxDB")
		}

		col := collector.NewCollector(ctx,
			appStorage,
			viper.GetString("collector.mqtt-broker"),
			dataStorage,
			fmt.Sprintf("%s:%d", viper.GetString("collector.address"), viper.GetInt("collector.port")))
		collectors, err := col.Start()
		if startError, ok := err.(collector.StartError); ok {
			ctx.WithError(startError).Warn("Could not start collecting all applications")
		} else if err != nil {
			ctx.WithError(err).Fatal("Could not start collector")
		}
		defer col.Stop()

		ctx.Infof("Started %d app collectors", len(collectors))

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
	},
}

func init() {
	RootCmd.AddCommand(collectorCmd)

	collectorCmd.Flags().String("address", "0.0.0.0", "The IP address to listen for management")
	collectorCmd.Flags().Int("port", 1783, "The port to listen for management")
	viper.BindPFlag("collector.address", collectorCmd.Flags().Lookup("address"))
	viper.BindPFlag("collector.port", collectorCmd.Flags().Lookup("port"))

	collectorCmd.Flags().String("mqtt-broker", "localhost:1883", "The address of the MQTT broker")
	viper.BindPFlag("collector.mqtt-broker", collectorCmd.Flags().Lookup("mqtt-broker"))

	collectorCmd.Flags().String("influxdb-addr", "http://localhost:8086", "The address of InfluxDB")
	collectorCmd.Flags().String("influxdb-username", "", "The username for InfluxDB")
	collectorCmd.Flags().String("influxdb-password", "", "The password for InfluxDB")
	viper.BindPFlag("collector.influxdb-addr", collectorCmd.Flags().Lookup("influxdb-addr"))
	viper.BindPFlag("collector.influxdb-username", collectorCmd.Flags().Lookup("influxdb-username"))
	viper.BindPFlag("collector.influxdb-password", collectorCmd.Flags().Lookup("influxdb-password"))

	collectorCmd.Flags().String("redis-addr", "localhost:6379", "The address of Redis")
	collectorCmd.Flags().Int("redis-db", 0, "The database of Redis")
	viper.BindPFlag("collector.redis-addr", collectorCmd.Flags().Lookup("redis-addr"))
	viper.BindPFlag("collector.redis-db", collectorCmd.Flags().Lookup("redis-db"))
}
