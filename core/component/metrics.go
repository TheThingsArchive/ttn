// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"strings"

	pb_protocol "github.com/TheThingsNetwork/api/protocol"
	"github.com/prometheus/client_golang/prometheus"
)

var receivedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "ttn",
		Name:      "messages_received_total",
		Help:      "Total number of uplink messages received.",
	}, []string{"message_type"},
)

var handledCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "ttn",
		Name:      "messages_handled_total",
		Help:      "Total number of uplink messages handled.",
	}, []string{"message_type"},
)

var handledBytes = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "ttn",
		Name:      "messages_handled_bytes_total",
		Help:      "Total number of message bytes handled.",
	}, []string{"message_type"},
)

func init() {
	prometheus.MustRegister(receivedCounter)
	prometheus.MustRegister(handledCounter)
	prometheus.MustRegister(handledBytes)
}

type message interface {
	GetPayload() []byte
	GetMessage() *pb_protocol.Message
}

func messageType(msg *pb_protocol.Message) string {
	if msg := msg.GetLoRaWAN(); msg != nil {
		mType := msg.GetMType().String()
		return strings.Replace(strings.Title(strings.ToLower(strings.Replace(mType, "_", " ", -1))), " ", "", -1)
	}
	return "Unknown"
}

func registerReceived(msg message) {
	receivedCounter.WithLabelValues(messageType(msg.GetMessage())).Inc()
}

func registerHandled(msg message) {
	mType := messageType(msg.GetMessage())
	handledCounter.WithLabelValues(mType).Inc()
	handledBytes.WithLabelValues(mType).Add(float64(len(msg.GetPayload())))
}

// RegisterReceived registers a received message
func (c *Component) RegisterReceived(msg message) { registerReceived(msg) }

// RegisterHandled registers a handled message
func (c *Component) RegisterHandled(msg message) { registerHandled(msg) }
