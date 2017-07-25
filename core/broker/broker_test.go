// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"sync"
	"testing"

	"github.com/TheThingsNetwork/ttn/utils/errors"

	pb "github.com/TheThingsNetwork/api/broker"
	pb_discovery "github.com/TheThingsNetwork/api/discovery"
	pb_handler "github.com/TheThingsNetwork/api/handler"
	pb_networkserver "github.com/TheThingsNetwork/api/networkserver"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	. "github.com/smartystreets/assertions"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
)

type mockHandlerDiscovery struct {
	a *pb_discovery.Announcement
}

func (d *mockHandlerDiscovery) ForAppID(appID string) (a []*pb_discovery.Announcement, err error) {
	return d.All()
}

func (d *mockHandlerDiscovery) Get(id string) (a *pb_discovery.Announcement, err error) {
	if d.a == nil {
		return nil, errors.New("Not found")
	}
	return d.a, nil
}

func (d *mockHandlerDiscovery) All() (a []*pb_discovery.Announcement, err error) {
	if d.a == nil {
		return []*pb_discovery.Announcement{}, nil
	}
	return []*pb_discovery.Announcement{d.a}, nil
}

func (d *mockHandlerDiscovery) AddAppID(_, _ string) error {
	return nil
}

type mockNetworkServer struct {
	devices []*pb_lorawan.Device
}

func (s *mockNetworkServer) GetDevices(ctx context.Context, req *pb_networkserver.DevicesRequest, options ...grpc.CallOption) (*pb_networkserver.DevicesResponse, error) {
	return &pb_networkserver.DevicesResponse{
		Results: s.devices,
	}, nil
}

func (s *mockNetworkServer) PrepareActivation(ctx context.Context, activation *pb.DeduplicatedDeviceActivationRequest, options ...grpc.CallOption) (*pb.DeduplicatedDeviceActivationRequest, error) {
	return activation, nil
}

func (s *mockNetworkServer) Activate(ctx context.Context, activation *pb_handler.DeviceActivationResponse, options ...grpc.CallOption) (*pb_handler.DeviceActivationResponse, error) {
	return activation, nil
}

func (s *mockNetworkServer) Uplink(ctx context.Context, message *pb.DeduplicatedUplinkMessage, options ...grpc.CallOption) (*pb.DeduplicatedUplinkMessage, error) {
	return message, nil
}

func (s *mockNetworkServer) Downlink(ctx context.Context, message *pb.DownlinkMessage, options ...grpc.CallOption) (*pb.DownlinkMessage, error) {
	return message, nil
}

func TestActivateDeactivateRouter(t *testing.T) {
	a := New(t)

	b := &broker{
		routers: make(map[string]chan *pb.DownlinkMessage),
	}

	err := b.DeactivateRouter("RouterID")
	a.So(err, ShouldNotBeNil)

	ch, err := b.ActivateRouter("RouterID")
	a.So(err, ShouldBeNil)
	a.So(ch, ShouldNotBeNil)

	_, err = b.ActivateRouter("RouterID")
	a.So(err, ShouldNotBeNil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for range ch {
		}
		wg.Done()
	}()

	err = b.DeactivateRouter("RouterID")
	a.So(err, ShouldBeNil)

	wg.Wait()
}

func TestActivateDeactivateHandler(t *testing.T) {
	a := New(t)

	b := &broker{
		handlers: make(map[string]*handler),
	}

	err := b.DeactivateHandlerUplink("HandlerID")
	a.So(err, ShouldNotBeNil)

	ch, err := b.ActivateHandlerUplink("HandlerID")
	a.So(err, ShouldBeNil)
	a.So(ch, ShouldNotBeNil)

	_, err = b.ActivateHandlerUplink("HandlerID")
	a.So(err, ShouldNotBeNil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for range ch {
		}
		wg.Done()
	}()

	err = b.DeactivateHandlerUplink("HandlerID")
	a.So(err, ShouldBeNil)

	wg.Wait()
}
