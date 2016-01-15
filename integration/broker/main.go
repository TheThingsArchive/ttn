// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
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
	// Parse options
	routersPort, handlersPort := parseOptions()

	// Create Logging Context
	log.SetHandler(text.New(os.Stdout))
	ctx := log.WithFields(log.Fields{
		"component": "Router",
	})

	// Instantiate all components
	rtrAdapter, err := http.NewAdapter(uint(routersPort), http.JSONPacketParser{}, ctx.WithField("tag", "Routers Adapter"))
	if err != nil {
		panic(err)
	}

	hdlHTTPAdapter, err := http.NewAdapter(uint(handlersPort), http.JSONPacketParser{}, ctx.WithField("tag", "Handlers Adapter"))
	if err != nil {
		panic(err)
	}

	hdlAdapter, err := pubsub.NewAdapter(hdlHTTPAdapter, pubsub.HandlerParser{}, ctx.WithField("tag", "Handlers Adapter"))
	if err != nil {
		panic(err)
	}

	broker, err := components.NewBroker(ctx.WithField("tag", "Broker"))
	if err != nil {
		panic(err)
	}

	// Bring the service to life

	// Listen to uplink
	go func() {
		for {
			packet, an, err := rtrAdapter.Next()
			if err != nil {
				fmt.Println(err)
				continue
			}
			go func(packet Packet, an AckNacker) {
				if err := broker.HandleUp(packet, an, hdlAdapter); err != nil {
					fmt.Println(err)
				}
			}(packet, an)
		}
	}()

	// List to handler registrations
	go func() {
		for {
			reg, an, err := hdlAdapter.NextRegistration()
			if err != nil {
				fmt.Println(err)
				continue
			}
			go func(reg Registration, an AckNacker) {
				if err := broker.Register(reg, an); err != nil {
					fmt.Println(err)
				}
			}(reg, an)
		}
	}()

	<-make(chan bool)
}

func parseOptions() (routersPort uint64, handlersPort uint64) {
	var routersPortFlag string
	var handlersPortFlag string
	flag.StringVar(&routersPortFlag, "routers-port", "", "TCP port on which the broker should listen to for incoming uplink packets.")
	flag.StringVar(&handlersPortFlag, "handlers-port", "", "TCP port on which the broker should listen to for incoming registrations and downlink packet.")
	flag.Parse()

	var err error
	routersPort, err = strconv.ParseUint(routersPortFlag, 10, 64)
	if err != nil {
		panic(err)
	}

	handlersPort, err = strconv.ParseUint(handlersPortFlag, 10, 64)
	if err != nil {
		panic(err)
	}

	return
}
