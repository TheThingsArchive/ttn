// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/hex"
	"fmt"
	"regexp"
)

// DeviceTopicType represents the type of a device topic
type DeviceTopicType string

const (
	// Activations of devices
	Activations DeviceTopicType = "activations"
	// Uplink data from devices
	Uplink DeviceTopicType = "up"
	// Downlink data to devices
	Downlink DeviceTopicType = "down"
)

// DeviceTopic represents an MQTT topic for application devices
// If the DevEUI is an empty []byte{}, it is considered to be a wildcard
type DeviceTopic struct {
	AppEUI []byte
	DevEUI []byte
	Type   DeviceTopicType
}

// ParseDeviceTopic parses an MQTT device topic string to a DeviceTopic struct
func ParseDeviceTopic(topic string) (*DeviceTopic, error) {
	pattern := regexp.MustCompile("([0-9A-F]{16}|\\+)/(devices)/([0-9A-F]{16}|\\+)/(activations|up|down)")
	matches := pattern.FindStringSubmatch(topic)

	if len(matches) < 4 {
		return nil, fmt.Errorf("Invalid topic format")
	}

	appEUI := []byte{}
	if matches[3] != "+" {
		appEUI, _ = hex.DecodeString(matches[1]) // validity asserted by our regex pattern
	}

	devEUI := []byte{}
	if matches[3] != "+" {
		devEUI, _ = hex.DecodeString(matches[3]) // validity asserted by our regex pattern
	}

	topicType := DeviceTopicType(matches[4])

	return &DeviceTopic{appEUI, devEUI, topicType}, nil
}

// String implements the Stringer interface
func (t DeviceTopic) String() string {
	appEUI := "+"
	if len(t.AppEUI) > 0 {
		appEUI = fmt.Sprintf("%X", t.AppEUI)
	}

	devEUI := "+"
	if len(t.DevEUI) > 0 {
		devEUI = fmt.Sprintf("%X", t.DevEUI)
	}
	return fmt.Sprintf("%s/%s/%s/%s", appEUI, "devices", devEUI, t.Type)
}
