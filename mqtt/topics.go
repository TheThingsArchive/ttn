// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/TheThingsNetwork/ttn/core/types"
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

const wildcard = "+"

// DeviceTopic represents an MQTT topic for application devices
// If the DevEUI is an empty []byte{}, it is considered to be a wildcard
type DeviceTopic struct {
	AppEUI types.AppEUI
	DevEUI types.DevEUI
	Type   DeviceTopicType
}

// ParseDeviceTopic parses an MQTT device topic string to a DeviceTopic struct
func ParseDeviceTopic(topic string) (*DeviceTopic, error) {
	var err error
	pattern := regexp.MustCompile("([0-9A-F]{16}|\\+)/(devices)/([0-9A-F]{16}|\\+)/(activations|up|down)")
	matches := pattern.FindStringSubmatch(topic)
	if len(matches) < 4 {
		return nil, fmt.Errorf("Invalid topic format")
	}
	var appEUI types.AppEUI
	if matches[1] != wildcard {
		appEUI, err = types.ParseAppEUI(matches[1])
		if err != nil {
			return nil, err
		}
	}
	var devEUI types.DevEUI
	if matches[1] != wildcard {
		devEUI, err = types.ParseDevEUI(matches[3])
		if err != nil {
			return nil, err
		}
	}
	topicType := DeviceTopicType(matches[4])
	return &DeviceTopic{appEUI, devEUI, topicType}, nil
}

// String implements the Stringer interface
func (t DeviceTopic) String() string {
	appEUI := wildcard
	if !reflect.DeepEqual(t.AppEUI, types.AppEUI{}) {
		appEUI = t.AppEUI.String()
	}
	devEUI := wildcard
	if !reflect.DeepEqual(t.DevEUI, types.DevEUI{}) {
		devEUI = t.DevEUI.String()
	}
	return fmt.Sprintf("%s/%s/%s/%s", appEUI, "devices", devEUI, t.Type)
}
