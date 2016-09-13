// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
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

const wildcard = "+"

// DeviceTopic represents an MQTT topic for application devices
type DeviceTopic struct {
	AppID string
	DevID string
	Type  DeviceTopicType
	Field string
}

// ParseDeviceTopic parses an MQTT device topic string to a DeviceTopic struct
func ParseDeviceTopic(topic string) (*DeviceTopic, error) {
	pattern := regexp.MustCompile("^([0-9a-z](?:[_-]?[0-9a-z]){1,35}|\\+)/(devices)/([0-9a-z](?:[_-]?[0-9a-z]){1,35}|\\+)/(activations|up|down)([0-9a-z/]+)?$")
	matches := pattern.FindStringSubmatch(topic)
	if len(matches) < 4 {
		return nil, fmt.Errorf("Invalid topic format")
	}
	var appID string
	if matches[1] != wildcard {
		appID = matches[1]
	}
	var devID string
	if matches[3] != wildcard {
		devID = matches[3]
	}
	topicType := DeviceTopicType(matches[4])
	deviceTopic := &DeviceTopic{appID, devID, topicType, ""}
	if topicType == Uplink && len(matches) > 4 {
		deviceTopic.Field = matches[5]
	}
	return deviceTopic, nil
}

// String implements the Stringer interface
func (t DeviceTopic) String() string {
	appID := wildcard
	if t.AppID != "" {
		appID = t.AppID
	}
	devID := wildcard
	if t.DevID != "" {
		devID = t.DevID
	}
	topic := fmt.Sprintf("%s/%s/%s/%s", appID, "devices", devID, t.Type)
	if t.Type == Uplink && t.Field != "" {
		topic += "/" + t.Field
	}
	return topic
}
