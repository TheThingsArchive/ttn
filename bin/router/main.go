// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package router/main creates and runs an working instance of a router on the host machine
package main

import (
	"flag"
	. "github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/adapters/gtw_rtr_udp"
	"github.com/thethingsnetwork/core/adapters/rtr_brk_http"
	"github.com/thethingsnetwork/core/components"
	"github.com/thethingsnetwork/core/utils/log"
	"strconv"
	"strings"
)

func main() {
	brokers, udpPort := parseOptions()
	router, err := components.NewRouter(brokers...)
	if err != nil {
		panic(err)
	}
	router.Logger = log.DebugLogger{Tag: "router"}

	upAdapter := gtw_rtr_udp.NewAdapter()
	upAdapter.Logger = log.DebugLogger{Tag: "upAdapter"}

	downAdapter := rtr_brk_http.NewAdapter()
	downAdapter.Logger = log.DebugLogger{Tag: "downAdapter"}

	router.Connect(upAdapter, downAdapter)
	if err := upAdapter.Listen(router, uint(udpPort)); err != nil {
		panic(err)
	}
	if err := downAdapter.Listen(router, brokers); err != nil {
		panic(err)
	}

	<-make(chan bool)
}

func parseOptions() (brokers []BrokerAddress, udpPort uint64) {
	var brokersFlag string
	var udpPortFlag string
	flag.StringVar(&brokersFlag, "brokers", "", `Broker addresses to which broadcast packets.
	For instance: 10.10.3.34:8080,thethingsnetwork.broker.com:3000
	`)
	flag.StringVar(&udpPortFlag, "udpPort", "", "Udp port on which the router should listen to.")
	flag.Parse()

	var err error
	udpPort, err = strconv.ParseUint(udpPortFlag, 10, 64)
	if err != nil {
		panic(err)
	}
	if brokersFlag == "" {
		panic("Need to provide at least one broker address")
	}

	brokersStr := strings.Split(brokersFlag, ",")
	for i := range brokersStr {
		brokers = append(brokers, BrokerAddress(strings.Trim(brokersStr[i], " ")))
	}
	return
}
