// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"errors"

	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type networkServerManager struct {
	*networkServer
}

var errf = grpc.Errorf

func (n *networkServerManager) getDevice(ctx context.Context, in *pb_lorawan.DeviceIdentifier) (*device.Device, error) {
	if in.AppId == "" || in.AppEui == nil || in.DevEui == nil {
		return nil, errf(codes.InvalidArgument, "AppID, AppEUI and DevEUI are required")
	}
	claims, err := n.Component.ValidateContext(ctx)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(in.AppId) {
		return nil, errf(codes.Unauthenticated, "No access to this device")
	}
	dev, err := n.devices.Get(*in.AppEui, *in.DevEui)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(dev.AppID) {
		return nil, errf(codes.Unauthenticated, "No access to this device")
	}
	return dev, nil
}

func (n *networkServerManager) GetDevice(ctx context.Context, in *pb_lorawan.DeviceIdentifier) (*pb_lorawan.Device, error) {
	dev, err := n.getDevice(ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb_lorawan.Device{
		AppId:            dev.AppID,
		AppEui:           &dev.AppEUI,
		DevEui:           &dev.DevEUI,
		DevAddr:          &dev.DevAddr,
		NwkSKey:          &dev.NwkSKey,
		FCntUp:           dev.FCntUp,
		FCntDown:         dev.FCntDown,
		DisableFCntCheck: dev.Options.DisableFCntCheck,
		Uses32BitFCnt:    dev.Options.Uses32BitFCnt,
	}, nil
}

func (n *networkServerManager) SetDevice(ctx context.Context, in *pb_lorawan.Device) (*api.Ack, error) {
	_, err := n.getDevice(ctx, &pb_lorawan.DeviceIdentifier{AppId: in.AppId, AppEui: in.AppEui, DevEui: in.DevEui})
	if err != nil && err != device.ErrNotFound {
		return nil, err
	}

	updated := &device.Device{
		AppID:    in.AppId,
		AppEUI:   *in.AppEui,
		DevEUI:   *in.DevEui,
		FCntUp:   in.FCntUp,
		FCntDown: in.FCntDown,
		Options: device.Options{
			DisableFCntCheck: in.DisableFCntCheck,
			Uses32BitFCnt:    in.Uses32BitFCnt,
		},
	}

	if in.NwkSKey != nil && in.DevAddr != nil {
		updated.DevAddr = *in.DevAddr
		updated.NwkSKey = *in.NwkSKey
	}

	err = n.devices.Set(updated)
	if err != nil {
		return nil, err
	}

	return &api.Ack{}, nil
}

func (n *networkServerManager) DeleteDevice(ctx context.Context, in *pb_lorawan.DeviceIdentifier) (*api.Ack, error) {
	_, err := n.getDevice(ctx, in)
	if err != nil {
		return nil, err
	}
	err = n.devices.Delete(*in.AppEui, *in.DevEui)
	if err != nil {
		return nil, err
	}
	return &api.Ack{}, nil
}

func (n *networkServerManager) GetStatus(ctx context.Context, in *pb.StatusRequest) (*pb.Status, error) {
	return nil, errors.New("Not Implemented")
}

// RegisterManager registers this networkserver as a NetworkServerManagerServer (github.com/TheThingsNetwork/ttn/api/networkserver)
func (n *networkServer) RegisterManager(s *grpc.Server) {
	server := &networkServerManager{n}
	pb.RegisterNetworkServerManagerServer(s, server)
	pb_lorawan.RegisterDeviceManagerServer(s, server)
}
