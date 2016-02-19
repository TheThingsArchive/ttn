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
	"strings"

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
				ctx.Errorf("Error building %s stat", name)
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
