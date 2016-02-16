// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/broadcast"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/parser"
	"github.com/TheThingsNetwork/ttn/core/adapters/semtech"
	"github.com/TheThingsNetwork/ttn/core/components"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
)

func main() {
	// Create Logging Context
	log.SetHandler(text.New(os.Stdout))
	log.SetLevel(log.DebugLevel)
	ctx := log.WithFields(log.Fields{
		"component": "Router",
	})

	// Parse options
	brokers, tcpPort, udpPort := parseOptions()

	// Instantiate all components
	gtwAdapter, err := semtech.NewAdapter(uint(udpPort), ctx.WithField("tag", "Gateway Adapter"))
	if err != nil {
		ctx.WithError(err).Fatal("Could not start Gateway Adapter")
	}

	pktAdapter, err := http.NewAdapter(uint(tcpPort), parser.JSON{}, ctx.WithField("tag", "Broker Adapter"))
	if err != nil {
		ctx.WithError(err).Fatal("Could not start Broker Adapter")
	}

	brkAdapter, err := broadcast.NewAdapter(pktAdapter, brokers, ctx.WithField("tag", "Broker Adapter"))
	if err != nil {
		ctx.WithError(err).Fatal("Could not start Broker Adapter")
	}

	db, err := components.NewRouterStorage(time.Hour * 8)
	if err != nil {
		ctx.WithError(err).Fatal("Could not create a local storage")
	}

	router := components.NewRouter(db, ctx.WithField("tag", "Router"))

	// Bring the service to life

	// Listen uplink
	go func() {
		for {
			packet, an, err := gtwAdapter.Next()
			if err != nil {
				ctx.WithError(err).Warn("Could not get next packet from gateway")
				continue
			}
			go func(packet Packet, an AckNacker) {
				if err := router.HandleUp(packet, an, brkAdapter); err != nil {
					ctx.WithError(err).Warn("Could not process packet from gateway")
				}
			}(packet, an)
		}
	}()

	// Listen broker registrations
	go func() {
		for {
			reg, an, err := brkAdapter.NextRegistration()
			if err != nil {
				ctx.WithError(err).Warn("Could not get next registration from broker")
				continue
			}
			go func(reg Registration, an AckNacker) {
				if err := router.Register(reg, an); err != nil {
					ctx.WithError(err).Warn("Could not process registration from broker")
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

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flags.StringVar(&brokersFlag, "brokers", "", `Broker addresses to which broadcast packets.
 	For instance: 10.10.3.34:8080,thethingsnetwork.broker.com:3000`)
	flags.StringVar(&udpPortFlag, "udp-port", "", "UDP port on which the router should listen to.")
	flags.StringVar(&tcpPortFlag, "tcp-port", "", "TCP port on which the router should listen to.")

	flags.Parse(os.Args[1:])

	var err error

	if tcpPortFlag == "" {
		log.Fatal("No TCP listen port supplied using the -tcp-port flag")
	}
	tcpPort, err = strconv.ParseUint(tcpPortFlag, 10, 64)
	if err != nil {
		log.Fatal("Could not parse the value for -tcp-port")
	}

	if udpPortFlag == "" {
		log.Fatal("No UDP listen port supplied using the -udp-port flag.")
	}
	udpPort, err = strconv.ParseUint(udpPortFlag, 10, 64)
	if err != nil {
		log.Fatal("Could not parse the value for -udp-port")
	}

	if brokersFlag == "" {
		log.Fatal("No broker address is supplied using -brokers flag.")
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
