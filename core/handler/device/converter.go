// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
)

// ToPb converts a device struct to its protocol buffer
func (d Device) ToPb() *pb_handler.Device {
	return &pb_handler.Device{
		AppId:       d.AppID,
		DevId:       d.DevID,
		Description: d.Description,
		Device:      &pb_handler.Device_LorawanDevice{LorawanDevice: d.ToLorawanPb()},
		Latitude:    d.Latitude,
		Longitude:   d.Longitude,
		Altitude:    d.Altitude,
		Attributes:  d.Attributes,
	}
}

// ToLorawanPb converts a device struct to a LoRaWAN protocol buffer
func (d Device) ToLorawanPb() *pb_lorawan.Device {
	return &pb_lorawan.Device{
		AppId:                 d.AppID,
		AppEui:                &d.AppEUI,
		DevId:                 d.DevID,
		DevEui:                &d.DevEUI,
		DevAddr:               &d.DevAddr,
		NwkSKey:               &d.NwkSKey,
		AppSKey:               &d.AppSKey,
		AppKey:                &d.AppKey,
		DisableFCntCheck:      d.Options.DisableFCntCheck,
		Uses32BitFCnt:         d.Options.Uses32BitFCnt,
		ActivationConstraints: d.Options.ActivationConstraints,
	}
}

// FromPb returns a new device from the given proto
func FromPb(in *pb_handler.Device) *Device {
	d := new(Device)
	d.FromPb(in)
	return d
}

// FromPb fills Device fields from a device proto
func (d *Device) FromPb(in *pb_handler.Device) {
	d.AppID = in.AppId
	d.DevID = in.DevId
	d.Description = in.Description
	d.Latitude = in.Latitude
	d.Longitude = in.Longitude
	d.Altitude = in.Altitude
	d.Attributes = in.Attributes
	d.FromLorawanPb(in.GetLorawanDevice())
}

// FromLorawanPb fills Device fields from a lorawan device proto
func (d *Device) FromLorawanPb(lorawan *pb_lorawan.Device) {
	if lorawan == nil {
		return
	}
	if lorawan.AppEui != nil {
		d.AppEUI = *lorawan.AppEui
	}
	if lorawan.DevEui != nil {
		d.DevEUI = *lorawan.DevEui
	}
	if lorawan.DevAddr != nil {
		d.DevAddr = *lorawan.DevAddr
	}
	if lorawan.AppKey != nil {
		d.AppKey = *lorawan.AppKey
	}
	if lorawan.AppSKey != nil {
		d.AppSKey = *lorawan.AppSKey
	}
	if lorawan.NwkSKey != nil {
		d.NwkSKey = *lorawan.NwkSKey
	}
	d.FCntUp = lorawan.FCntUp
	d.Options = Options{
		DisableFCntCheck:      lorawan.DisableFCntCheck,
		Uses32BitFCnt:         lorawan.Uses32BitFCnt,
		ActivationConstraints: lorawan.ActivationConstraints,
	}
}
