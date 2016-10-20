// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"regexp"
	"strings"
)

const simpleWildcard = "+"
const wildcard = "#"

// DeviceTopicType represents the type of a device topic
type DeviceTopicType string

// Topic types for Devices
const (
	DeviceEvents   DeviceTopicType = "events"
	DeviceUplink   DeviceTopicType = "up"
	DeviceDownlink DeviceTopicType = "down"
)

// DeviceTopic represents an MQTT topic for devices
type DeviceTopic struct {
	AppID string
	DevID string
	Type  DeviceTopicType
	Field string
}

// ParseDeviceTopic parses an MQTT device topic string to a DeviceTopic struct
func ParseDeviceTopic(topic string) (*DeviceTopic, error) {
	pattern := regexp.MustCompile("^([0-9a-z](?:[_-]?[0-9a-z]){1,35}|\\+)/(devices)/([0-9a-z](?:[_-]?[0-9a-z]){1,35}|\\+)/(events|up|down)([0-9a-z/]+)?$")
	matches := pattern.FindStringSubmatch(topic)
	if len(matches) < 4 {
		return nil, fmt.Errorf("Invalid topic format")
	}
	var appID string
	if matches[1] != simpleWildcard {
		appID = matches[1]
	}
	var devID string
	if matches[3] != simpleWildcard {
		devID = matches[3]
	}
	topicType := DeviceTopicType(matches[4])
	deviceTopic := &DeviceTopic{appID, devID, topicType, ""}
	if (topicType == DeviceUplink || topicType == DeviceEvents) && len(matches) > 4 {
		deviceTopic.Field = strings.Trim(matches[5], "/")
	}
	return deviceTopic, nil
}

// String implements the Stringer interface
func (t DeviceTopic) String() string {
	appID := simpleWildcard
	if t.AppID != "" {
		appID = t.AppID
	}
	devID := simpleWildcard
	if t.DevID != "" {
		devID = t.DevID
	}
	if t.Type == DeviceEvents && t.Field == "" {
		t.Field = simpleWildcard
	}
	topic := fmt.Sprintf("%s/%s/%s/%s", appID, "devices", devID, t.Type)
	if (t.Type == DeviceUplink || t.Type == DeviceEvents) && t.Field != "" {
		topic += "/" + t.Field
	}
	return topic
}

// ApplicationTopicType represents the type of an application topic
type ApplicationTopicType string

// Topic types for Applications
const (
	AppEvents ApplicationTopicType = "events"
)

// ApplicationTopic represents an MQTT topic for applications
type ApplicationTopic struct {
	AppID string
	Type  ApplicationTopicType
	Field string
}

// ParseApplicationTopic parses an MQTT device topic string to an ApplicationTopic struct
func ParseApplicationTopic(topic string) (*ApplicationTopic, error) {
	pattern := regexp.MustCompile("^([0-9a-z](?:[_-]?[0-9a-z]){1,35}|\\+)/(events)([0-9a-z/-]+|/#)?$")
	matches := pattern.FindStringSubmatch(topic)
	if len(matches) < 2 {
		return nil, fmt.Errorf("Invalid topic format")
	}
	var appID string
	if matches[1] != simpleWildcard {
		appID = matches[1]
	}
	topicType := ApplicationTopicType(matches[2])
	appTopic := &ApplicationTopic{appID, topicType, ""}
	if topicType == AppEvents && len(matches) > 2 {
		appTopic.Field = strings.Trim(matches[3], "/")
	}
	return appTopic, nil
}

// String implements the Stringer interface
func (t ApplicationTopic) String() string {
	appID := simpleWildcard
	if t.AppID != "" {
		appID = t.AppID
	}
	if t.Type == AppEvents && t.Field == "" {
		t.Field = wildcard
	}
	topic := fmt.Sprintf("%s/%s", appID, t.Type)
	if t.Type == AppEvents && t.Field != "" {
		topic += "/" + t.Field
	}
	return topic
}
