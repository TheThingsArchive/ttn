// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package statuspage extends the basic http to show statistics on GET /status
//
// The adapter registers a new endpoint [/status/] to an original basic http adapter and serves statistics on it.
package statuspage

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	httpadapter "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/apex/log"
	"github.com/rcrowley/go-metrics"
)

// Adapter extending the default http adapter
type Adapter struct {
	*httpadapter.Adapter               // The original http adapter
	ctx                  log.Interface // Logging context
}

// NewAdapter constructs a new http adapter that also handles status requests
func NewAdapter(adapter *httpadapter.Adapter, ctx log.Interface) (*Adapter, error) {
	a := &Adapter{
		Adapter: adapter,
		ctx:     ctx,
	}

	a.RegisterEndpoint("/status/", a.handleStatus)

	return a, nil
}

// handle request [GET] on /status
func (a *Adapter) handleStatus(w http.ResponseWriter, req *http.Request) {
	ctx := a.ctx.WithField("sender", req.RemoteAddr)
	ctx.Debug("Receiving new status request")

	// Check the http method
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Unreckognized HTTP method. Please use [GET] to get the status"))
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
