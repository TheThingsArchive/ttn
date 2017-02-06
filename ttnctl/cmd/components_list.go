// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

func plural(n int, name string) string {
	if n == 1 {
		return name
	}
	return fmt.Sprintf("%ss", name)
}

var componentsListCmd = &cobra.Command{
	Use:   "list [Type]",
	Short: "Get the token for a network component.",
	Long:  `components token gets a singed token for the component.`,
	Example: `$ ttnctld components list                                                                                                                                                         146 !
  INFO Found 0 routers
  INFO Found 0 brokers
  INFO Found 1 handler

 	Type   	ID
1	handler	test
`,
	Run: func(cmd *cobra.Command, args []string) {
		assertArgsLength(cmd, args, 0, 1)

		act := util.GetAccount(ctx)

		components, err := act.ListComponents()
		if err != nil {
			ctx.WithError(err).Fatal("Could get list network components")
		}

		typ := ""
		if len(args) == 1 {
			typ = args[0]
		}

		routers := make([]account.Component, 0)
		handlers := make([]account.Component, 0)
		brokers := make([]account.Component, 0)

		for _, component := range components {
			switch component.Type {
			case string(account.Handler):
				handlers = append(handlers, component)
			case string(account.Broker):
				brokers = append(brokers, component)
			case string(account.Router):
				routers = append(routers, component)
			}
		}

		if typ == "" || typ == "routers" || typ == "router" {
			ctx.Info(fmt.Sprintf("Found %v %s", len(routers), plural(len(routers), "router")))

			if len(routers) > 0 {
				table := uitable.New()
				table.MaxColWidth = 70
				table.AddRow("", "Type", "ID")
				for i, router := range routers {
					table.AddRow(i, router.Type, router.ID)
				}

				fmt.Println()
				fmt.Println(table)
				fmt.Println()
			}
		}

		if typ == "" || typ == "brokers" || typ == "broker" {
			ctx.Info(fmt.Sprintf("Found %v %s", len(brokers), plural(len(brokers), "broker")))

			if len(brokers) > 0 {
				table := uitable.New()
				table.MaxColWidth = 70
				table.AddRow("", "Type", "ID")
				for i, broker := range brokers {
					table.AddRow(i, broker.Type, broker.ID)
				}

				fmt.Println()
				fmt.Println(table)
				fmt.Println()
			}
		}

		if typ == "" || typ == "handlers" || typ == "handler" {
			ctx.Info(fmt.Sprintf("Found %v %s", len(handlers), plural(len(handlers), "handler")))

			if len(handlers) > 0 {
				table := uitable.New()
				table.MaxColWidth = 70
				table.AddRow("", "Type", "ID")
				for i, handler := range handlers {
					table.AddRow(i, handler.Type, handler.ID)
				}

				fmt.Println()
				fmt.Println(table)
				fmt.Println()
			}
		}
	},
}

func init() {
	componentsCmd.AddCommand(componentsListCmd)
}
