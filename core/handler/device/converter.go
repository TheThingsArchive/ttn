// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	pb_handler "github.com/TheThingsNetwork/api/handler"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
)

// ToPb converts a device struct to its protocol buffer
func (d Device) ToPb() *pb_handler.Device {
	return &pb_handler.Device{
		AppID:       d.AppID,
		DevID:       d.DevID,
		Description: d.Description,
		Device:      &pb_handler.Device_LoRaWANDevice{LoRaWANDevice: d.ToLoRaWANPb()},
		Latitude:    d.Latitude,
		Longitude:   d.Longitude,
		Altitude:    d.Altitude,
		Attributes:  d.Attributes,
	}
}

// ToLoRaWANPb converts a device struct to a LoRaWAN protocol buffer
func (d Device) ToLoRaWANPb() *pb_lorawan.Device {
	return &pb_lorawan.Device{
		AppID:                 d.AppID,
		AppEUI:                &d.AppEUI,
		DevID:                 d.DevID,
		DevEUI:                &d.DevEUI,
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
	d.AppID = in.AppID
	d.DevID = in.DevID
	d.Description = in.Description
	d.Latitude = in.Latitude
	d.Longitude = in.Longitude
	d.Altitude = in.Altitude
	d.Attributes = in.Attributes
	d.FromLoRaWANPb(in.GetLoRaWANDevice())
}

// FromLoRaWANPb fills Device fields from a lorawan device proto
func (d *Device) FromLoRaWANPb(lorawan *pb_lorawan.Device) {
	if lorawan == nil {
		return
	}
	if lorawan.AppEUI != nil {
		d.AppEUI = *lorawan.AppEUI
	}
	if lorawan.DevEUI != nil {
		d.DevEUI = *lorawan.DevEUI
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
