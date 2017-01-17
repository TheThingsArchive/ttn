// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"fmt"
	"regexp"
	"strings"
)

const simpleWildcard = "*"
const wildard = "#"

// DeviceKeyType represents the type of a device topic
type DeviceKeyType string

// Topic types for Devices
const (
	DeviceEvents   DeviceKeyType = "events"
	DeviceUplink   DeviceKeyType = "up"
	DeviceDownlink DeviceKeyType = "down"
)

// DeviceKey represents an AMQP routing key for devices
type DeviceKey struct {
	AppID string
	DevID string
	Type  DeviceKeyType
	Field string
}

// ParseDeviceKey parses an AMQP device routing key string to a DeviceKey struct
func ParseDeviceKey(key string) (*DeviceKey, error) {
	pattern := regexp.MustCompile("^([0-9a-z](?:[_-]?[0-9a-z]){1,35}|\\*)\\.(devices)\\.([0-9a-z](?:[_-]?[0-9a-z]){1,35}|\\*)\\.(events|up|down)([0-9a-z\\.]+)?$")
	matches := pattern.FindStringSubmatch(key)
	if len(matches) < 5 {
		return nil, fmt.Errorf("Invalid key format")
	}
	var appID string
	if matches[1] != simpleWildcard {
		appID = matches[1]
	}
	var devID string
	if matches[3] != simpleWildcard {
		devID = matches[3]
	}
	keyType := DeviceKeyType(matches[4])
	deviceKey := &DeviceKey{appID, devID, keyType, ""}
	if keyType == DeviceEvents && len(matches) > 5 {
		deviceKey.Field = strings.Trim(matches[5], ".")
	}
	return deviceKey, nil
}

// String implements the Stringer interface
func (t DeviceKey) String() string {
	appID := simpleWildcard
	if t.AppID != "" {
		appID = t.AppID
	}
	devID := simpleWildcard
	if t.DevID != "" {
		devID = t.DevID
	}
	key := fmt.Sprintf("%s.%s.%s.%s", appID, "devices", devID, t.Type)
	if t.Type == DeviceEvents && t.Field != "" {
		key += "." + t.Field
	}
	return key
}

// ApplicationKeyType represents an AMQP application routing key
type ApplicationKeyType string

// Topic types for Applications
const (
	AppEvents ApplicationKeyType = "events"
)

// ApplicationKey represents an AMQP topic for applications
type ApplicationKey struct {
	AppID string
	Type  ApplicationKeyType
	Field string
}

// ParseApplicationKey parses an AMQP application routing key string to an ApplicationKey struct
func ParseApplicationKey(key string) (*ApplicationKey, error) {
	pattern := regexp.MustCompile("^([0-9a-z](?:[_-]?[0-9a-z]){1,35}|\\*)\\.(events)([0-9a-z\\.-]+|\\.#)?$")
	matches := pattern.FindStringSubmatch(key)
	if len(matches) < 3 {
		return nil, fmt.Errorf("Invalid key format")
	}
	var appID string
	if matches[1] != simpleWildcard {
		appID = matches[1]
	}
	keyType := ApplicationKeyType(matches[2])
	appKey := &ApplicationKey{appID, keyType, ""}
	if keyType == AppEvents && len(matches) > 3 {
		appKey.Field = strings.Trim(matches[3], ".")
	}
	return appKey, nil
}

// String implements the Stringer interface
func (t ApplicationKey) String() string {
	appID := simpleWildcard
	if t.AppID != "" {
		appID = t.AppID
	}
	if t.Type == AppEvents && t.Field == "" {
		t.Field = wildard
	}
	key := fmt.Sprintf("%s.%s", appID, t.Type)
	if t.Type == AppEvents && t.Field != "" {
		key += "." + t.Field
	}
	return key
}
