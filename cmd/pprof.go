// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// +build dev

package cmd

import (
	"bytes"
	"io/ioutil"
	_ "net/http/pprof"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/TheThingsNetwork/ttn/api/stats"
	"github.com/spf13/cobra"
)

func init() {
	cobra.OnInitialize(initPprof)
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
}
