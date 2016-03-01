// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"flag"
	"os"
	"strings"

	"github.com/TheThingsNetwork/ttn/simulators/gateway"
	"github.com/TheThingsNetwork/ttn/simulators/node"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
)

var (
	numNodes int
	interval int
)

func main() {
	routers := parseOptions()

	log.SetHandler(text.New(os.Stdout))
	log.SetLevel(log.DebugLevel)

	nodeCtx := log.WithFields(log.Fields{"Simulator": "Node"})
	gatewayCtx := log.WithFields(log.Fields{"Simulator": "Gateway"})

	nodes := []node.LiveNode{}

	for i := 0; i < numNodes; i++ {
		nodes = append(nodes, node.New(interval, nodeCtx))
	}

	gateway.MockRandomly(nodes, gatewayCtx, routers...)
}

func parseOptions() (routers []string) {
	var routersFlag string
	flag.StringVar(&routersFlag, "routers", "", `Router addresses to which send packets.
 	For instance: 10.10.3.34:8080,thethingsnetwork.router.com:3000
 	`)

	flag.IntVar(&interval, "interval", 500, "Average time (in milliseconds) between node messages")
	flag.IntVar(&numNodes, "nodes", 10, "Number of nodes to simulate")

	flag.Parse()

	if routersFlag == "" {
		panic("Need to provide at least one router address")
	}

	routers = strings.Split(routersFlag, ",")
	return
}
