// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"flag"
	"strings"

	"github.com/TheThingsNetwork/ttn/simulators/gateway"
	"github.com/TheThingsNetwork/ttn/simulators/node"
)

func main() {
	routers := parseOptions()

	nodes := []node.LiveNode{
		node.New(),
		node.New(),
		node.New(),
		node.New(),
		node.New(),
		node.New(),
		node.New(),
		node.New(),
		node.New(),
		node.New(),
		node.New(),
		node.New(),
	}

	gateway.MockRandomly(nodes, routers...)
}

func parseOptions() (routers []string) {
	var routersFlag string
	flag.StringVar(&routersFlag, "routers", "", `Router addresses to which send packets.
 	For instance: 10.10.3.34:8080,thethingsnetwork.router.com:3000
 	`)
	flag.Parse()

	if routersFlag == "" {
		panic("Need to provide at least one router address")
	}

	routers = strings.Split(routersFlag, ",")
	return
}
