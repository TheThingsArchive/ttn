// Copyright © 2017 The Things Industries B.V.

package handler

import (
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	hdl "github.com/TheThingsNetwork/ttn/core/handler/device"
)

//Discard old device and change it with a device fetch from the handler/device.go#Device
func DevFromHdl(dev *hdl.Device) *Device {
	return &Device{
		AppId:       dev.AppID,
		DevId:       dev.DevID,
		Description: dev.Description,
		Device: &Device_LorawanDevice{LorawanDevice: &pb_lorawan.Device{
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

func DevToHdl(dev *hdl.Device, in *Device, lorawan *pb_lorawan.Device) {

	dev.AppID = in.AppId
	dev.AppEUI = *lorawan.AppEui
	dev.DevID = in.DevId
	dev.DevEUI = *lorawan.DevEui
	dev.Description = in.Description
	dev.Latitude = in.Latitude
	dev.Longitude = in.Longitude
	dev.Altitude = in.Altitude
	fromLorawan(dev, lorawan)
	dev.Attributes = in.Attributes
}

func fromLorawan(dev *hdl.Device, lorawan *pb_lorawan.Device) {
	if lorawan.DevAddr != nil {
		dev.DevAddr = *lorawan.DevAddr
	}
	if lorawan.NwkSKey != nil {
		dev.NwkSKey = *lorawan.NwkSKey
	}
	if lorawan.AppSKey != nil {
		dev.AppSKey = *lorawan.AppSKey
	}
	if lorawan.AppKey != nil {
		if dev.AppKey != *lorawan.AppKey { // When the AppKey of an existing device is changed
			dev.UsedAppNonces = []hdl.AppNonce{}
			dev.UsedDevNonces = []hdl.DevNonce{}
		}
		dev.AppKey = *lorawan.AppKey
	}
	dev.Options = hdl.Options{
		DisableFCntCheck:      lorawan.DisableFCntCheck,
		Uses32BitFCnt:         lorawan.Uses32BitFCnt,
		ActivationConstraints: lorawan.ActivationConstraints,
	}
	if dev.Options.ActivationConstraints == "" {
		dev.Options.ActivationConstraints = "local"
	}
}
