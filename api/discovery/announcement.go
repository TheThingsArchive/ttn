// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"encoding/json"
	"fmt"
)

// AnnouncementProperties contains all properties of an Announcement that can
// be stored in Redis.
var AnnouncementProperties = []string{
	"id",
	"token",
	"description",
	"service_name",
	"service_version",
	"net_address",
	"metadata",
}

// ToStringStringMap converts the given properties of Announcement to a
// map[string]string for storage in Redis.
func (announcement *Announcement) ToStringStringMap(properties ...string) (map[string]string, error) {
	output := make(map[string]string)
	for _, p := range properties {
		property, err := announcement.formatProperty(p)
		if err != nil {
			return output, err
		}
		if property != "" {
			output[p] = property
		}
	}
	return output, nil
}

// FromStringStringMap imports known values from the input to a Status.
func (announcement *Announcement) FromStringStringMap(input map[string]string) error {
	for k, v := range input {
		announcement.parseProperty(k, v)
	}
	return nil
}

func (announcement *Announcement) formatProperty(property string) (formatted string, err error) {
	switch property {
	case "id":
		formatted = announcement.Id
	case "token":
		formatted = announcement.Token
	case "description":
		formatted = announcement.Description
	case "service_name":
		formatted = announcement.ServiceName
	case "service_version":
		formatted = announcement.ServiceVersion
	case "net_address":
		formatted = announcement.NetAddress
	case "metadata":
		json, err := json.Marshal(announcement.Metadata)
		if err != nil {
			return "", err
		}
		formatted = string(json)
	default:
		err = fmt.Errorf("Property %s does not exist in Announcement", property)
	}
	return
}

func (announcement *Announcement) parseProperty(property string, value string) error {
	if value == "" {
		return nil
	}
	switch property {
	case "id":
		announcement.Id = value
	case "token":
		announcement.Token = value
	case "description":
		announcement.Description = value
	case "service_name":
		announcement.ServiceName = value
	case "service_version":
		announcement.ServiceVersion = value
	case "net_address":
		announcement.NetAddress = value
	case "metadata":
		metadata := []*Metadata{}
		err := json.Unmarshal([]byte(value), &metadata)
		if err != nil {
			return err
		}
		announcement.Metadata = metadata
	default:
		return fmt.Errorf("Property %s does not exist in Announcement", property)
	}
	return nil
}
