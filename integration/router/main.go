// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/broadcast"
	"github.com/TheThingsNetwork/ttn/core/adapters/semtech"
	"github.com/TheThingsNetwork/ttn/core/components"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
)

func main() {
	// Parse options
	brokers, tcpPort, udpPort := parseOptions()

	// Create Logging Context
	log.SetHandler(text.New(os.Stdout))
	log.SetLevel(log.DebugLevel)
	ctx := log.WithFields(log.Fields{
		"component": "Router",
	})

	// Instantiate all components
	gtwAdapter, err := semtech.NewAdapter(uint(udpPort), ctx.WithField("tag", "Gateway Adapter"))
	if err != nil {
		panic(err)
	}

	pktAdapter, err := http.NewAdapter(uint(tcpPort), http.JSONPacketParser{}, ctx.WithField("tag", "Broker Adapter"))
	if err != nil {
		panic(err)
	}

	brkAdapter, err := broadcast.NewAdapter(pktAdapter, brokers, ctx.WithField("tag", "Broker Adapter"))
	if err != nil {
		panic(err)
	}

	router, err := components.NewRouter(ctx.WithField("tag", "Router"))
	if err != nil {
		panic(err)
	}

	// Bring the service to life

	// Listen uplink
	go func() {
		for {
			packet, an, err := gtwAdapter.Next()
			if err != nil {
				fmt.Println(err)
				continue
			}
			go func(packet Packet, an AckNacker) {
				if err := router.HandleUp(packet, an, brkAdapter); err != nil {
					fmt.Println(err)
				}
			}(packet, an)
		}
	}()

	// Listen broker registrations
	go func() {
		for {
			reg, an, err := brkAdapter.NextRegistration()
			if err != nil {
				fmt.Println(err)
				continue
			}
			go func(reg Registration, an AckNacker) {
				if err := router.Register(reg, an); err != nil {
					fmt.Println(err)
				}
			}(reg, an)
		}
	}()

	<-make(chan bool)
}

func parseOptions() (brokers []Recipient, tcpPort uint64, udpPort uint64) {
	var brokersFlag string
	var udpPortFlag string
	var tcpPortFlag string
	flag.StringVar(&brokersFlag, "brokers", "", `Broker addresses to which broadcast packets.
 	For instance: 10.10.3.34:8080,thethingsnetwork.broker.com:3000
 	`)
	flag.StringVar(&udpPortFlag, "udp-port", "", "UDP port on which the router should listen to.")
	flag.StringVar(&tcpPortFlag, "tcp-port", "", "TCP port on which the router should listen to.")
	flag.Parse()

	var err error
	tcpPort, err = strconv.ParseUint(tcpPortFlag, 10, 64)
	if err != nil {
		panic(err)
	}
	udpPort, err = strconv.ParseUint(udpPortFlag, 10, 64)
	if err != nil {
		panic(err)
	}
	if brokersFlag == "" {
		panic("Need to provide at least one broker address")
	}

	brokersStr := strings.Split(brokersFlag, ",")
	for i := range brokersStr {
		brokers = append(brokers, Recipient{
			Address: strings.Trim(brokersStr[i], " "),
			Id:      i,
		})

	}
	return
}
