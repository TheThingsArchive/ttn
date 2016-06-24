// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/core/storage"
)

// StatusMessageProperties contains all properties of a StatusMessage that can
// be stored in Redis.
var StatusMessageProperties = []string{
	"timestamp",
	"time",
	"ip",
	"platform",
	"contact_email",
	"description",
	"region",
	"gps.time",
	"gps.latitude",
	"gps.longitude",
	"gps.altitude",
	"rtt",
	"rx_in",
	"rx_ok",
	"tx_in",
	"tx_ok",
}

// ToStringStringMap converts the given properties of Status to a
// map[string]string for storage in Redis.
func (status *Status) ToStringStringMap(properties ...string) (map[string]string, error) {
	output := make(map[string]string)
	for _, p := range properties {
		property, err := status.formatProperty(p)
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
func (status *Status) FromStringStringMap(input map[string]string) error {
	for k, v := range input {
		status.parseProperty(k, v)
	}
	return nil
}

func (status *Status) formatProperty(property string) (formatted string, err error) {
	switch property {
	case "timestamp":
		formatted = storage.FormatUint32(status.Timestamp)
	case "time":
		formatted = storage.FormatInt64(status.Time)
	case "ip":
		formatted = strings.Join(status.Ip, ",")
	case "platform":
		formatted = status.Platform
	case "contact_email":
		formatted = status.ContactEmail
	case "description":
		formatted = status.Description
	case "region":
		formatted = status.Region
	case "gps.time":
		if status.Gps != nil {
			formatted = storage.FormatInt64(status.Gps.Time)
		}
	case "gps.latitude":
		if status.Gps != nil {
			formatted = storage.FormatFloat32(status.Gps.Latitude)
		}
	case "gps.longitude":
		if status.Gps != nil {
			formatted = storage.FormatFloat32(status.Gps.Longitude)
		}
	case "gps.altitude":
		if status.Gps != nil {
			formatted = storage.FormatInt32(status.Gps.Altitude)
		}
	case "rtt":
		formatted = storage.FormatUint32(status.Rtt)
	case "rx_in":
		formatted = storage.FormatUint32(status.RxIn)
	case "rx_ok":
		formatted = storage.FormatUint32(status.RxOk)
	case "tx_in":
		formatted = storage.FormatUint32(status.TxIn)
	case "tx_ok":
		formatted = storage.FormatUint32(status.TxOk)
	default:
		err = fmt.Errorf("Property %s does not exist in Status", property)
	}
	return
}

func (status *Status) parseProperty(property string, value string) error {
	if value == "" {
		return nil
	}
	switch property {
	case "timestamp":
		val, err := storage.ParseUint32(value)
		if err != nil {
			return err
		}
		status.Timestamp = val
	case "time":
		val, err := storage.ParseInt64(value)
		if err != nil {
			return err
		}
		status.Time = val
	case "ip":
		val := strings.Split(value, ",")
		status.Ip = val
	case "platform":
		status.Platform = value
	case "contact_email":
		status.ContactEmail = value
	case "description":
		status.Description = value
	case "region":
		status.Region = value
	case "gps.time":
		if status.Gps == nil {
			status.Gps = &GPSMetadata{}
		}
		val, err := storage.ParseInt64(value)
		if err != nil {
			return err
		}
		status.Gps.Time = val
	case "gps.latitude":
		if status.Gps == nil {
			status.Gps = &GPSMetadata{}
		}
		val, err := storage.ParseFloat32(value)
		if err != nil {
			return err
		}
		status.Gps.Latitude = val
	case "gps.longitude":
		if status.Gps == nil {
			status.Gps = &GPSMetadata{}
		}
		val, err := storage.ParseFloat32(value)
		if err != nil {
			return err
		}
		status.Gps.Longitude = val
	case "gps.altitude":
		if status.Gps == nil {
			status.Gps = &GPSMetadata{}
		}
		val, err := storage.ParseInt32(value)
		if err != nil {
			return err
		}
		status.Gps.Altitude = val
	case "rtt":
		val, err := storage.ParseUint32(value)
		if err != nil {
			return err
		}
		status.Rtt = val
	case "rx_in":
		val, err := storage.ParseUint32(value)
		if err != nil {
			return err
		}
		status.RxIn = val
	case "rx_ok":
		val, err := storage.ParseUint32(value)
		if err != nil {
			return err
		}
		status.RxOk = val
	case "tx_in":
		val, err := storage.ParseUint32(value)
		if err != nil {
			return err
		}
		status.TxIn = val
	case "tx_ok":
		val, err := storage.ParseUint32(value)
		if err != nil {
			return err
		}
		status.TxOk = val
	}
	return nil
}
