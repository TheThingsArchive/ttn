// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

//+build ignore

package main

import (
	"crypto/tls"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/ttn/api/monitor"
	"github.com/TheThingsNetwork/ttn/api/monitor/monitorserver"
	"google.golang.org/grpc"
)

func main() {
	log.Set(apex.Stdout())
	ctx := log.Get()

	if len(os.Args) != 2 {
		ctx.Fatal("Usage: ttn-monitor-server-example [listen]")
	}

	var tlsConfig tls.Config
	certificate, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		ctx.WithError(err).Fatal("Could not load tls certificate and key")
	}
	tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
	lis, err := tls.Listen("tcp", os.Args[1], &tlsConfig)
	if err != nil {
		ctx.WithError(err).Fatal("Failed to listen")
	}
	s := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint16))
	server := monitorserver.NewReferenceMonitorServer(10)
	monitor.RegisterMonitorServer(s, server)
	go s.Serve(lis)
	ctx.Infof("Listening on %s", lis.Addr().String())

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	ctx.WithField("signal", <-sigChan).Info("signal received")

	s.Stop()
}
