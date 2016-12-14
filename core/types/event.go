// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

// EventType represents the type of event
type EventType string

// Event types
const (
	UplinkErrorEvent       EventType = "up/errors"
	DownlinkScheduledEvent EventType = "down/scheduled"
	DownlinkSentEvent      EventType = "down/sent"
	DownlinkErrorEvent     EventType = "down/errors"
	DownlinkAckEvent       EventType = "down/acks"
	ActivationEvent        EventType = "activations"
	ActivationErrorEvent   EventType = "activations/errors"
)

// DeviceEvent represents an application-layer event message for a device event
type DeviceEvent struct {
	AppID string
	DevID string
	Event EventType
	Data  interface{}
}

// ErrorEventData is added to error events
type ErrorEventData struct {
	Error string `json:"error"`
}

// ActivationEventData is added to activation events
type ActivationEventData struct {
	AppEUI   AppEUI   `json:"app_eui"`
	DevEUI   DevEUI   `json:"dev_eui"`
	DevAddr  DevAddr  `json:"dev_addr"`
	Metadata Metadata `json:"metadata"`
}

// DownlinkEventConfigInfo contains configuration information for a downlink message, all fields are optional
type DownlinkEventConfigInfo struct {
	Modulation string `json:"modulation,omitempty"`
	DataRate   string `json:"data_rate,omitempty"`
	BitRate    uint   `json:"bit_rate,omitempty"`
	FCnt       uint   `json:"counter,omitempty"`
	Frequency  uint   `json:"frequency,omitempty"`
	Power      int    `json:"power,omitempty"`
}

// DownlinkEventData is added to downlink events
type DownlinkEventData struct {
	Payload   []byte                  `json:"payload"`
	GatewayID string                  `json:"gateway_id"`
	Config    DownlinkEventConfigInfo `json:"config"`
}
