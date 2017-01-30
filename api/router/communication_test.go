// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func newTestRouter() *testRouter {
	return &testRouter{
		RouterStreamServer: NewRouterStreamServer(),
	}
}

type testRouter struct {
	*RouterStreamServer
}

func (s *testRouter) Activate(context.Context, *DeviceActivationRequest) (*DeviceActivationResponse, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "Not implemented")
}

func (s *testRouter) Serve(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	RegisterRouterServer(srv, s)
	srv.Serve(lis)
}

func TestRouterCommunication(t *testing.T) {
	a := New(t)

	ctx := GetLogger(t, "TestRouterCommunication")
	log.Set(ctx)

	rtr := newTestRouter()
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(1000) + 10000
	go rtr.Serve(port)

	conn, _ := api.Dial(fmt.Sprintf("localhost:%d", port))

	{
		rtr.UplinkChanFunc = func(md metadata.MD) (chan *UplinkMessage, error) {
			ch := make(chan *UplinkMessage, 1)
			go func() {
				ctx.Info("[SERVER] Channel opened")
				for message := range ch {
					ctx.WithField("Message", message).Info("[SERVER] Received Uplink")
				}
				ctx.Info("[SERVER] Channel closed")
			}()
			return ch, nil
		}

		rtrClient := NewRouterClient(conn)
		gtwClient := NewRouterClientForGateway(rtrClient, "dev", "token")
		uplink := NewMonitoredUplinkStream(gtwClient)

		err := uplink.Send(&UplinkMessage{
			Payload: []byte{1, 2, 3, 4},
			ProtocolMetadata: &protocol.RxMetadata{Protocol: &protocol.RxMetadata_Lorawan{Lorawan: &lorawan.Metadata{
				Modulation: lorawan.Modulation_LORA,
				DataRate:   "SF7BW125",
				CodingRate: "4/7",
			}}},
			GatewayMetadata: &gateway.RxMetadata{
				GatewayId: "dev",
			},
		})

		a.So(err, ShouldBeNil)

		time.Sleep(10 * time.Millisecond)

		uplink.Close()

		time.Sleep(10 * time.Millisecond)

		gtwClient.Close()

		time.Sleep(10 * time.Millisecond)
	}

	{
		rtr.GatewayStatusChanFunc = func(md metadata.MD) (chan *gateway.Status, error) {
			ch := make(chan *gateway.Status, 1)
			go func() {
				ctx.Info("[SERVER] Channel opened")
				for message := range ch {
					ctx.WithField("Message", message).Info("[SERVER] Received GatewayStatus")
				}
				ctx.Info("[SERVER] Channel closed")
			}()
			return ch, nil
		}

		rtrClient := NewRouterClient(conn)
		gtwClient := NewRouterClientForGateway(rtrClient, "dev", "token")
		status := NewMonitoredGatewayStatusStream(gtwClient)

		err := status.Send(&gateway.Status{Time: time.Now().UnixNano()})

		a.So(err, ShouldBeNil)

		time.Sleep(10 * time.Millisecond)

		status.Close()

		time.Sleep(10 * time.Millisecond)

		gtwClient.Close()

		time.Sleep(10 * time.Millisecond)
	}

	{
		rtr.DownlinkChanFunc = func(md metadata.MD) (<-chan *DownlinkMessage, func(), error) {
			ch := make(chan *DownlinkMessage, 1)
			stop := make(chan struct{})
			cancel := func() {
				ctx.Info("[SERVER] Canceling downlink")
				close(stop)
			}
			go func() {
			loop:
				for {
					select {
					case <-stop:
						break loop
					case <-time.After(5 * time.Millisecond):
						ctx.Info("[SERVER] Sending Downlink")
						ch <- &DownlinkMessage{
							Payload: []byte{1, 2, 3, 4},
						}
					}
				}
				close(ch)
				ctx.Info("[SERVER] Closed Downlink")
			}()
			return ch, cancel, nil
		}

		rtrClient := NewRouterClient(conn)
		gtwClient := NewRouterClientForGateway(rtrClient, "dev", "token")
		downlink := NewMonitoredDownlinkStream(gtwClient)

		ch := downlink.Channel()

		go func() {
			for downlink := range ch {
				ctx.WithField("Downlink", downlink).Info("[CLIENT] Received Downlink")
			}
			ctx.Info("[CLIENT] Closed Downlink")
		}()

		time.Sleep(10 * time.Millisecond)

		downlink.Close()

		time.Sleep(10 * time.Millisecond)

		gtwClient.Close()

		time.Sleep(10 * time.Millisecond)
	}

}
