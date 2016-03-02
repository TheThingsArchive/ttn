// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	. "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/rcrowley/go-metrics"
)

// StatusPage shows statistic on GEt request
//
// It listens to requests of the form: [GET] /status/
//
// No body or query param are expected
//
// This handler does not generate any registration.
type StatusPage struct{}

// URL implements the http.Handler interface
func (p StatusPage) URL() string {
	return "/status/"
}

// Handle implements the http.Handler interface
func (p StatusPage) Handle(w http.ResponseWriter, chpkt chan<- PktReq, chreg chan<- RegReq, req *http.Request) {
	// Check the http method
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Unreckognized HTTP method. Please use [GET] to transfer a packet"))
		return
	}

	remoteHost, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		//The HTTP server did not set RemoteAddr to IP:port, which would be very very strange.
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	remoteIP := net.ParseIP(remoteHost)
	if remoteIP == nil || !remoteIP.IsLoopback() {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("Status is only available from the local host, not from %s", remoteIP)))
		return
	}

	allStats := make(map[string]interface{})

	metrics.Each(func(name string, i interface{}) {
		// Make sure we put things in the right place
		thisStat := allStats
		for _, path := range strings.Split(name, ".") {
			if thisStat[path] == nil {
				thisStat[path] = make(map[string]interface{})
			}
			if _, ok := thisStat[path].(map[string]interface{}); ok {
				thisStat = thisStat[path].(map[string]interface{})
			} else {
				return
			}
		}

		switch metric := i.(type) {

		case metrics.Counter:
			m := metric.Snapshot()
			thisStat["count"] = m.Count()

		case metrics.Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles([]float64{0.25, 0.5, 0.75})

			thisStat["avg"] = h.Mean()
			thisStat["min"] = h.Min()
			thisStat["max"] = h.Max()
			thisStat["p_25"] = ps[0]
			thisStat["p_50"] = ps[1]
			thisStat["p_75"] = ps[2]

		case metrics.Meter:
			m := metric.Snapshot()

			thisStat["rate_1"] = m.Rate1()
			thisStat["rate_5"] = m.Rate5()
			thisStat["rate_15"] = m.Rate15()
			thisStat["count"] = m.Count()

		}
	})

	response, err := json.Marshal(allStats)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
