// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/TheThingsNetwork/ttn/utils/security"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// genkeysCmd represents the genkeys command
func genkeysCmd(component string) *cobra.Command {
	return &cobra.Command{
		Use:   "genkeys",
		Short: "Generate keys and certificate",
		Long:  `ttn genkeys generates keys and a TLS certificate for this component`,
		Run: func(cmd *cobra.Command, args []string) {
			var names []string
			names = append(names, viper.GetString(component+".server-address-announce"))
			names = append(names, args...)
			err := security.GenerateKeys(viper.GetString("key-dir"), names...)
			if err != nil {
				ctx.WithError(err).Fatal("Could not generate keys")
			}
			ctx.WithField("TLSDir", viper.GetString("key-dir")).Info("Done")
		},
	}
}

func init() {
	routerCmd.AddCommand(genkeysCmd("router"))
	brokerCmd.AddCommand(genkeysCmd("broker"))
	handlerCmd.AddCommand(genkeysCmd("handler"))
	networkserverCmd.AddCommand(genkeysCmd("networkserver"))
}
