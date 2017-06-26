// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// +build dev

package cmd

import (
	"bytes"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/TheThingsNetwork/ttn/api/stats"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(initPprof)

	RootCmd.PersistentFlags().String("pprof-address", "", "The address to start listening on for pprof debugging")
}

func initPprof() {
	var buf bytes.Buffer
	go func() {
		for {
			time.Sleep(time.Minute)

			runtime.GC()
			if err := pprof.WriteHeapProfile(&buf); err != nil {
				if ctx != nil {
					ctx.Warnf("could not write memory profile: %s", err)
				}
			}
			ioutil.WriteFile("/tmp/ttn.mem.prof", buf.Bytes(), 0644)
			buf.Reset()

			ctx.WithField("ComponentStats", stats.GetComponent()).Debug("Stats")
		}
	}()

	go func() {
		time.Sleep(time.Second)
		addr := viper.GetString("pprof-address")
		if addr != "" {
			ctx.WithField("Address", addr).Debug("Adding pprof server")
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				ctx.WithError(err).Warnf("Could not start pprof server")
			}
		}
	}()
}
