// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

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

// Url implements the http.Handler interface
func (p StatusPage) Url() string {
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

	allStats := statData{}

	metrics.Each(func(name string, i interface{}) {
		switch metric := i.(type) {

		case metrics.Counter:
			m := metric.Snapshot()
			allStats[name] = m.Count()

		case metrics.Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles([]float64{0.25, 0.5, 0.75})

			allStats[name] = histData{
				Avg: h.Mean(),
				Min: h.Min(),
				Max: h.Max(),
				P25: ps[0],
				P50: ps[1],
				P75: ps[2],
			}

		case metrics.Meter:
			m := metric.Snapshot()
			allStats[name] = meterData{
				Rate1:  m.Rate1(),
				Rate5:  m.Rate5(),
				Rate15: m.Rate15(),
				Count:  m.Count(),
			}
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

type statData map[string]interface{}

type meterData struct {
	Rate1  float64 `json:"rate_1"`
	Rate5  float64 `json:"rate_5"`
	Rate15 float64 `json:"rate_15"`
	Count  int64   `json:"count"`
}

type histData struct {
	Avg float64 `json:"avg"`
	Min int64   `json:"min"`
	Max int64   `json:"max"`
	P25 float64 `json:"p_25"`
	P50 float64 `json:"p_50"`
	P75 float64 `json:"p_75"`
}
