// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"errors"
	"fmt"
	"strings"
)

// TopicType represents the type of a topic
type TopicType string

const (
	// Devices indicates a topic for devices
	Devices TopicType = "devices"
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

// DeviceTopic represents a publish/subscribe topic for application devices
type DeviceTopic struct {
	AppEUI string
	DevEUI string
	Type   DeviceTopicType
}

// GetTopicType returns the type of the specified topic
func GetTopicType(topic string) (TopicType, error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 2 {
		return "", errors.New("Invalid format")
	}
	return TopicType(parts[1]), nil
}

// DecodeDeviceTopic decodes the specified topic in a DeviceTopic structure
func DecodeDeviceTopic(topic string) (*DeviceTopic, error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 4 {
		return nil, errors.New("Invalid format")
	}
	if TopicType(parts[1]) != Devices {
		return nil, errors.New("Not a device topic")
	}

	return &DeviceTopic{parts[0], parts[2], DeviceTopicType(parts[3])}, nil
}

// Encode encodes the DeviceTopic to a topic
func (t *DeviceTopic) Encode() (string, error) {
	return fmt.Sprintf("%s/%s/%s/%s", t.AppEUI, Devices, t.DevEUI, t.Type), nil
}
