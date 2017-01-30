// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	pb "github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/utils/fcnt"
)

func (n *networkServer) HandleGetDevices(req *pb.DevicesRequest) (*pb.DevicesResponse, error) {
	devices, err := n.devices.ListForAddress(*req.DevAddr)
	if err != nil {
		return nil, err
	}

	// Return all devices with DevAddr with FCnt <= fCnt or Security off

	res := &pb.DevicesResponse{
		Results: make([]*pb_lorawan.Device, 0, len(devices)),
	}

	for _, device := range devices {
		if device == nil {
			continue
		}
		fullFCnt := fcnt.GetFull(device.FCntUp, uint16(req.FCnt))
		dev := &pb_lorawan.Device{
			AppEui:           &device.AppEUI,
			AppId:            device.AppID,
			DevEui:           &device.DevEUI,
			DevId:            device.DevID,
			NwkSKey:          &device.NwkSKey,
			FCntUp:           device.FCntUp,
			Uses32BitFCnt:    device.Options.Uses32BitFCnt,
			DisableFCntCheck: device.Options.DisableFCntCheck,
		}
		if device.Options.DisableFCntCheck {
			res.Results = append(res.Results, dev)
			continue
		}
		if device.FCntUp <= req.FCnt {
			res.Results = append(res.Results, dev)
			continue
		} else if device.Options.Uses32BitFCnt && device.FCntUp <= fullFCnt {
			res.Results = append(res.Results, dev)
			continue
		}
	}

	return res, nil
}
