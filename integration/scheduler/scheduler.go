// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"flag"
	"strings"

	"github.com/TheThingsNetwork/ttn/simulators/gateway"
	"time"
)

func main() {
	routers, delay, schedule := parseOptions()
	gateway.MockWithSchedule(schedule, delay, routers...)
}

func parseOptions() (routers []string, delay time.Duration, schedule string) {
	var routersFlag string
	var delayFlag string
	flag.StringVar(&routersFlag, "routers", "", `Router addresses to which send packets.
 	For instance: 10.10.3.34:8080,thethingsnetwork.router.com:3000
 	`)
	flag.StringVar(&delayFlag, "delay", "", `Interval of time between 2 sending.
	For instance: 500ms
	`)
	flag.StringVar(&schedule, "schedule", "", "JSON file defining the packets to schedule")
	flag.Parse()

	var err error
	if delay, err = time.ParseDuration(delayFlag); err != nil {
		panic(err)
	}

	if routersFlag == "" {
		panic("Need to provide at least one router address")
	}

	routers = strings.Split(routersFlag, ",")
	return
}
