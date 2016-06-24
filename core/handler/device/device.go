// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
)

// Options for the specified device
type Options struct {
	DisableFCntCheck bool // Disable Frame counter check (insecure)
	Uses32BitFCnt    bool // Use 32-bit Frame counters
}

type DevNonce [2]byte
type AppNonce [3]byte

// Device contains the state of a device
type Device struct {
	DevEUI        types.DevEUI
	AppEUI        types.AppEUI
	AppID         string
	DevAddr       types.DevAddr
	AppKey        types.AppKey
	UsedDevNonces []DevNonce
	UsedAppNonces []AppNonce
	NwkSKey       types.NwkSKey
	AppSKey       types.AppSKey
	NextDownlink  *mqtt.DownlinkMessage
}

// DeviceProperties contains all properties of a Device that can be stored in Redis.
var DeviceProperties = []string{
	"dev_eui",
	"app_eui",
	"app_id",
	"dev_addr",
	"app_key",
	"nwk_s_key",
	"app_s_key",
	"used_dev_nonces",
	"used_app_nonces",
	"next_downlink",
}

// ToStringStringMap converts the given properties of Device to a
// map[string]string for storage in Redis.
func (device *Device) ToStringStringMap(properties ...string) (map[string]string, error) {
	output := make(map[string]string)
	for _, p := range properties {
		property, err := device.formatProperty(p)
		if err != nil {
			return output, err
		}
		if property != "" {
			output[p] = property
		}
	}
	return output, nil
}

// FromStringStringMap imports known values from the input to a Device.
func (device *Device) FromStringStringMap(input map[string]string) error {
	for k, v := range input {
		device.parseProperty(k, v)
	}
	return nil
}

func (device *Device) formatProperty(property string) (formatted string, err error) {
	switch property {
	case "dev_eui":
		formatted = device.DevEUI.String()
	case "app_eui":
		formatted = device.AppEUI.String()
	case "app_id":
		formatted = device.AppID
	case "dev_addr":
		formatted = device.DevAddr.String()
	case "app_key":
		formatted = device.AppKey.String()
	case "nwk_s_key":
		formatted = device.NwkSKey.String()
	case "app_s_key":
		formatted = device.AppSKey.String()
	case "used_dev_nonces":
		nonces := make([]string, 0, len(device.UsedDevNonces))
		for _, nonce := range device.UsedDevNonces {
			nonces = append(nonces, fmt.Sprintf("%X", nonce))
		}
		formatted = strings.Join(nonces, ",")
	case "used_app_nonces":
		nonces := make([]string, 0, len(device.UsedAppNonces))
		for _, nonce := range device.UsedAppNonces {
			nonces = append(nonces, fmt.Sprintf("%X", nonce))
		}
		formatted = strings.Join(nonces, ",")
	case "next_downlink":
		var jsonBytes []byte
		jsonBytes, err = json.Marshal(device.NextDownlink)
		formatted = string(jsonBytes)
	default:
		err = fmt.Errorf("Property %s does not exist in Device", property)
	}
	return
}

func (device *Device) parseProperty(property string, value string) error {
	if value == "" {
		return nil
	}
	switch property {
	case "dev_eui":
		val, err := types.ParseDevEUI(value)
		if err != nil {
			return err
		}
		device.DevEUI = val
	case "app_eui":
		val, err := types.ParseAppEUI(value)
		if err != nil {
			return err
		}
		device.AppEUI = val
	case "app_id":
		device.AppID = value
	case "dev_addr":
		val, err := types.ParseDevAddr(value)
		if err != nil {
			return err
		}
		device.DevAddr = val
	case "app_key":
		val, err := types.ParseAppKey(value)
		if err != nil {
			return err
		}
		device.AppKey = val
	case "nwk_s_key":
		val, err := types.ParseNwkSKey(value)
		if err != nil {
			return err
		}
		device.NwkSKey = val
	case "app_s_key":
		val, err := types.ParseAppSKey(value)
		if err != nil {
			return err
		}
		device.AppSKey = val
	case "used_dev_nonces":
		nonceStrs := strings.Split(value, ",")
		nonces := make([]DevNonce, 0, len(nonceStrs))
		for _, nonceStr := range nonceStrs {
			var nonce DevNonce
			nonceBytes, err := types.ParseHEX(nonceStr, 2)
			if err != nil {
				return err
			}
			copy(nonce[:], nonceBytes)
			nonces = append(nonces, nonce)
		}
		device.UsedDevNonces = nonces
	case "used_app_nonces":
		nonceStrs := strings.Split(value, ",")
		nonces := make([]AppNonce, 0, len(nonceStrs))
		for _, nonceStr := range nonceStrs {
			var nonce AppNonce
			nonceBytes, err := types.ParseHEX(nonceStr, 3)
			if err != nil {
				return err
			}
			copy(nonce[:], nonceBytes)
			nonces = append(nonces, nonce)
		}
		device.UsedAppNonces = nonces
	case "next_downlink":
		if value != "null" {
			nextDownlink := &mqtt.DownlinkMessage{}
			err := json.Unmarshal([]byte(value), nextDownlink)
			if err != nil {
				return err
			}
			device.NextDownlink = nextDownlink
		}
	}
	return nil
}
