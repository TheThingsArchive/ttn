// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"github.com/prometheus/client_golang/prometheus"
)

var duplicatesHistogram = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "ttn",
		Subsystem: "broker",
		Name:      "message_duplicates",
		Help:      "Histogram of message duplicates.",
		Buckets:   []float64{0, 1, 2, 4, 8, 16, 32, 64, 128},
	},
)

var micChecksHistogram = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "ttn",
		Subsystem: "broker",
		Name:      "mic_checks",
		Help:      "Histogram of MIC checks.",
		Buckets:   []float64{0, 1, 2, 4, 8, 16, 32, 64, 128},
	},
)

var connectedRouters = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "ttn",
		Subsystem: "broker",
		Name:      "connected_routers",
		Help:      "Number of connected routers.",
	},
)

var connectedHandlers = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "ttn",
		Subsystem: "broker",
		Name:      "connected_handlers",
		Help:      "Number of connected handlers.",
	},
)

var initialized = false

func initMetrics() {
	if initialized {
		return
	}
	initialized = true
	prometheus.MustRegister(duplicatesHistogram)
	prometheus.MustRegister(micChecksHistogram)
	prometheus.MustRegister(connectedRouters)
	prometheus.MustRegister(connectedHandlers)
}
