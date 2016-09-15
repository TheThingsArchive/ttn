// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"fmt"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type networkServerManager struct {
	networkServer *networkServer
}

func (n *networkServerManager) getDevice(ctx context.Context, in *pb_lorawan.DeviceIdentifier) (*device.Device, error) {
	if !in.Validate() {
		return nil, errors.NewErrInvalidArgument("Device Identifier", "validation failed")
	}
	claims, err := n.networkServer.Component.ValidateTTNAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	dev, err := n.networkServer.devices.Get(*in.AppEui, *in.DevEui)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(dev.AppID) {
		return nil, errors.NewErrPermissionDenied(fmt.Sprintf("No access to Application %s", dev.AppID))
	}
	return dev, nil
}

func (n *networkServerManager) GetDevice(ctx context.Context, in *pb_lorawan.DeviceIdentifier) (*pb_lorawan.Device, error) {
	dev, err := n.getDevice(ctx, in)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}

	lastSeen := time.Unix(0, 0)
	if !dev.LastSeen.IsZero() {
		lastSeen = dev.LastSeen
	}

	return &pb_lorawan.Device{
		AppId:            dev.AppID,
		AppEui:           &dev.AppEUI,
		DevId:            dev.DevID,
		DevEui:           &dev.DevEUI,
		DevAddr:          &dev.DevAddr,
		NwkSKey:          &dev.NwkSKey,
		FCntUp:           dev.FCntUp,
		FCntDown:         dev.FCntDown,
		DisableFCntCheck: dev.Options.DisableFCntCheck,
		Uses32BitFCnt:    dev.Options.Uses32BitFCnt,
		LastSeen:         lastSeen.UnixNano(),
	}, nil
}

func (n *networkServerManager) SetDevice(ctx context.Context, in *pb_lorawan.Device) (*empty.Empty, error) {
	_, err := n.getDevice(ctx, &pb_lorawan.DeviceIdentifier{AppEui: in.AppEui, DevEui: in.DevEui})
	if err != nil && errors.GetErrType(err) != errors.NotFound {
		return nil, errors.BuildGRPCError(err)
	}

	if !in.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Device Identifier")
	}

	updated := &device.Device{
		AppID:    in.AppId,
		AppEUI:   *in.AppEui,
		DevID:    in.DevId,
		DevEUI:   *in.DevEui,
		FCntUp:   in.FCntUp,
		FCntDown: in.FCntDown,
		Options: device.Options{
			DisableFCntCheck:      in.DisableFCntCheck,
			Uses32BitFCnt:         in.Uses32BitFCnt,
			ActivationConstraints: in.ActivationConstraints,
		},
	}

	if in.NwkSKey != nil && in.DevAddr != nil {
		updated.DevAddr = *in.DevAddr
		updated.NwkSKey = *in.NwkSKey
	}

	err = n.networkServer.devices.Set(updated)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}

	return &empty.Empty{}, nil
}

func (n *networkServerManager) DeleteDevice(ctx context.Context, in *pb_lorawan.DeviceIdentifier) (*empty.Empty, error) {
	_, err := n.getDevice(ctx, in)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	err = n.networkServer.devices.Delete(*in.AppEui, *in.DevEui)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	return &empty.Empty{}, nil
}

func (n *networkServerManager) GetPrefixes(ctx context.Context, in *pb_lorawan.PrefixesRequest) (*pb_lorawan.PrefixesResponse, error) {
	var mapping []*pb_lorawan.PrefixesResponse_PrefixMapping
	for prefix, usage := range n.networkServer.prefixes {
		mapping = append(mapping, &pb_lorawan.PrefixesResponse_PrefixMapping{
			Prefix: prefix.String(),
			Usage:  usage,
		})
	}
	return &pb_lorawan.PrefixesResponse{
		Prefixes: mapping,
	}, nil
}

func (n *networkServerManager) GetDevAddr(ctx context.Context, in *pb_lorawan.DevAddrRequest) (*pb_lorawan.DevAddrResponse, error) {
	devAddr, err := n.networkServer.getDevAddr(in.Usage...)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	return &pb_lorawan.DevAddrResponse{
		DevAddr: &devAddr,
	}, nil
}

func (n *networkServerManager) GetStatus(ctx context.Context, in *pb.StatusRequest) (*pb.Status, error) {
	return nil, grpcErrf(codes.Unimplemented, "Not Implemented")
}

// RegisterManager registers this networkserver as a NetworkServerManagerServer (github.com/TheThingsNetwork/ttn/api/networkserver)
func (n *networkServer) RegisterManager(s *grpc.Server) {
	server := &networkServerManager{n}
	pb.RegisterNetworkServerManagerServer(s, server)
	pb_lorawan.RegisterDeviceManagerServer(s, server)
	pb_lorawan.RegisterDevAddrManagerServer(s, server)
}
