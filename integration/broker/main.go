// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"flag"
	"os"
	"strconv"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/core/adapters/http/pubsub"
	"github.com/TheThingsNetwork/ttn/core/components"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
)

func main() {
	// Create Logging Context
	log.SetHandler(text.New(os.Stdout))
	log.SetLevel(log.DebugLevel)
	ctx := log.WithFields(log.Fields{
		"component": "Broker",
	})

	// Parse options
	routersPort, handlersPort := parseOptions()

	// Instantiate all components
	rtrAdapter, err := http.NewAdapter(uint(routersPort), http.JSONPacketParser{}, ctx.WithField("tag", "Routers Adapter"))
	if err != nil {
		ctx.WithError(err).Fatal("Could not start Routers Adapter")
	}

	hdlHTTPAdapter, err := http.NewAdapter(uint(handlersPort), http.JSONPacketParser{}, ctx.WithField("tag", "Handlers Adapter"))
	if err != nil {
		ctx.WithError(err).Fatal("Could not start Handlers Adapter")
	}

	hdlAdapter, err := pubsub.NewAdapter(hdlHTTPAdapter, pubsub.HandlerParser{}, ctx.WithField("tag", "Handlers Adapter"))
	if err != nil {
		ctx.WithError(err).Fatal("Could not start Handlers Adapter")
	}

	broker, err := components.NewBroker(ctx.WithField("tag", "Broker"))
	if err != nil {
		ctx.WithError(err).Fatal("Could not start Broker")
	}

	// Bring the service to life

	// Listen to uplink
	go func() {
		for {
			packet, an, err := rtrAdapter.Next()
			if err != nil {
				ctx.WithError(err).Error("Could not retrieve uplink")
				continue
			}
			go func(packet Packet, an AckNacker) {
				if err := broker.HandleUp(packet, an, hdlAdapter); err != nil {
					ctx.WithError(err).Error("Could not process uplink")
				}
			}(packet, an)
		}
	}()

	// List to handler registrations
	go func() {
		for {
			reg, an, err := hdlAdapter.NextRegistration()
			if err != nil {
				ctx.WithError(err).Error("Could not retrieve registration")
				continue
			}
			go func(reg Registration, an AckNacker) {
				if err := broker.Register(reg, an); err != nil {
					ctx.WithError(err).Error("Could not process registration")
				}
			}(reg, an)
		}
	}()

	<-make(chan bool)
}

func parseOptions() (routersPort uint64, handlersPort uint64) {
	var routersPortFlag string
	var handlersPortFlag string

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flags.StringVar(&routersPortFlag, "routers-port", "", "TCP port on which the broker should listen to for incoming uplink packets.")
	flags.StringVar(&handlersPortFlag, "handlers-port", "", "TCP port on which the broker should listen to for incoming registrations and downlink packet.")

	flags.Parse(os.Args[1:])

	var err error

	if routersPortFlag == "" {
		log.Fatal("No Router listen port supplied using the -routers-port flag")
	}
	routersPort, err = strconv.ParseUint(routersPortFlag, 10, 64)
	if err != nil {
		log.Fatal("Could not parse the value for -routers-port")
	}

	if handlersPortFlag == "" {
		log.Fatal("No Handler listen port supplied using the -handlers-port flag")
	}
	handlersPort, err = strconv.ParseUint(handlersPortFlag, 10, 64)
	if err != nil {
		log.Fatal("Could not parse the value for -handlers-port")
	}

	return
}
