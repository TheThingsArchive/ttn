// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// handlerCmd represents the handler command
var handlerCmd = &cobra.Command{
	Use:   "handler",
	Short: "The Things Network handler",
	Long: `
The default handler is the bridge between The Things Network and applications.
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		ctx = ctx.WithField("cmd", "handler")
		ctx.WithFields(log.Fields{
			"database":     viper.GetString("handler.database"),
			"brokers-port": viper.GetInt("handler.brokers-port"),
		}).Info("Using Configuration")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx.Fatal("The handler's not ready yet, but we're working on it!")

		// TODO: Add logic
	},
}

func init() {
	RootCmd.AddCommand(handlerCmd)

	handlerCmd.Flags().String("database", "boltdb:/tmp/ttn_handler.db", "Database connection")
	handlerCmd.Flags().Int("brokers-port", 1691, "TCP port for connections from brokers")

	viper.BindPFlag("handler.database", handlerCmd.Flags().Lookup("database"))
	viper.BindPFlag("handler.brokers-port", handlerCmd.Flags().Lookup("brokers-port"))

}
