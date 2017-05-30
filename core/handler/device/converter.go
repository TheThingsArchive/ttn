// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
)

// Convert a device representation to a protoBuffer device struct
func (dev *Device) ToPb() *pb_handler.Device {
	return &pb_handler.Device{
		AppId:       dev.AppID,
		DevId:       dev.DevID,
		Description: dev.Description,
		Device: &pb_handler.Device_LorawanDevice{LorawanDevice: &pb_lorawan.Device{
			AppId:                 dev.AppID,
			AppEui:                &dev.AppEUI,
			DevId:                 dev.DevID,
			DevEui:                &dev.DevEUI,
			DevAddr:               &dev.DevAddr,
			NwkSKey:               &dev.NwkSKey,
			AppSKey:               &dev.AppSKey,
			AppKey:                &dev.AppKey,
			DisableFCntCheck:      dev.Options.DisableFCntCheck,
			Uses32BitFCnt:         dev.Options.Uses32BitFCnt,
			ActivationConstraints: dev.Options.ActivationConstraints,
		}},
		Latitude:   dev.Latitude,
		Longitude:  dev.Longitude,
		Altitude:   dev.Altitude,
		Attributes: dev.Attributes,
	}
}

// FromPb creates a new Device from a protoBuffer Device
func (dev *Device) FromPb(in *pb_handler.Device, lorawan *pb_lorawan.Device) *Device {
	dev.AppID = in.AppId
	dev.AppEUI = *lorawan.AppEui
	dev.DevID = in.DevId
	dev.DevEUI = *lorawan.DevEui
	dev.Description = in.Description
	dev.Latitude = in.Latitude
	dev.Longitude = in.Longitude
	dev.Altitude = in.Altitude
	dev.fromLorawan(lorawan)
	dev.Attributes = in.Attributes
	return dev
}

// fromLorawan fill a device with lorawan device infos
func (dev *Device) fromLorawan(lorawan *pb_lorawan.Device) {
	if lorawan.DevAddr != nil {
		dev.DevAddr = *lorawan.DevAddr
	}
	if lorawan.NwkSKey != nil {
		dev.NwkSKey = *lorawan.NwkSKey
	}
	if lorawan.AppSKey != nil {
		dev.AppSKey = *lorawan.AppSKey
	}
	dev.Options = Options{
		DisableFCntCheck:      lorawan.DisableFCntCheck,
		Uses32BitFCnt:         lorawan.Uses32BitFCnt,
		ActivationConstraints: lorawan.ActivationConstraints,
	}
	if dev.Options.ActivationConstraints == "" {
		dev.Options.ActivationConstraints = "local"
	}
}
