// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func newTestBroker() *testBroker {
	return &testBroker{
		BrokerStreamServer: NewBrokerStreamServer(),
	}
}

type testBroker struct {
	*BrokerStreamServer
}

var _ BrokerServer = &testBroker{}

func (s *testBroker) Activate(context.Context, *DeviceActivationRequest) (*DeviceActivationResponse, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "Not implemented")
}

func (s *testBroker) Serve(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	RegisterBrokerServer(srv, s)
	srv.Serve(lis)
}

func TestHandlerBrokerCommunication(t *testing.T) {
	a := New(t)

	ctx := GetLogger(t, "TestHandlerBrokerCommunication")
	log.Set(ctx)

	brk := newTestBroker()
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(1000) + 10000
	go brk.Serve(port)

	conn, _ := api.Dial(fmt.Sprintf("localhost:%d", port))

	{
		brk.HandlerPublishChanFunc = func(md metadata.MD) (chan *DownlinkMessage, error) {
			ch := make(chan *DownlinkMessage, 1)
			go func() {
				ctx.Info("[SERVER] Channel opened")
				for message := range ch {
					ctx.WithField("Message", message).Info("[SERVER] Received Downlink")
				}
				ctx.Info("[SERVER] Channel closed")
			}()
			return ch, nil
		}

		brkClient := NewBrokerClient(conn)
		downlink := NewMonitoredHandlerPublishStream(brkClient, func() context.Context {
			return context.Background()
		})

		err := downlink.Send(&DownlinkMessage{
			Payload: []byte{1, 2, 3, 4},
		})

		a.So(err, ShouldBeNil)

		time.Sleep(10 * time.Millisecond)

		downlink.Close()

		time.Sleep(10 * time.Millisecond)
	}

	{
		brk.HandlerSubscribeChanFunc = func(md metadata.MD) (<-chan *DeduplicatedUplinkMessage, func(), error) {
			ch := make(chan *DeduplicatedUplinkMessage, 1)
			stop := make(chan struct{})
			cancel := func() {
				ctx.Info("[SERVER] Canceling uplink")
				close(stop)
			}
			go func() {
			loop:
				for {
					select {
					case <-stop:
						break loop
					case <-time.After(5 * time.Millisecond):
						ctx.Info("[SERVER] Sending Uplink")
						ch <- &DeduplicatedUplinkMessage{
							Payload: []byte{1, 2, 3, 4},
						}
					}
				}
				close(ch)
				ctx.Info("[SERVER] Closed Uplink")
			}()
			return ch, cancel, nil
		}

		brkClient := NewBrokerClient(conn)
		uplink := NewMonitoredHandlerSubscribeStream(brkClient, func() context.Context {
			return context.Background()
		})

		ch := uplink.Channel()

		go func() {
			for uplink := range ch {
				ctx.WithField("Uplink", uplink).Info("[CLIENT] Received Uplink")
			}
			ctx.Info("[CLIENT] Closed Uplink")
		}()

		time.Sleep(10 * time.Millisecond)

		uplink.Close()

		time.Sleep(10 * time.Millisecond)
	}

}

func TestRouterBrokerCommunication(t *testing.T) {
	a := New(t)

	ctx := GetLogger(t, "TestRouterBrokerCommunication")
	log.Set(ctx)

	brk := newTestBroker()
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(1000) + 10000
	go brk.Serve(port)

	conn, _ := api.Dial(fmt.Sprintf("localhost:%d", port))

	{
		brk.RouterAssociateChanFunc = func(md metadata.MD) (chan *UplinkMessage, <-chan *DownlinkMessage, func(), error) {
			up := make(chan *UplinkMessage, 1)
			down := make(chan *DownlinkMessage, 1)

			stop := make(chan struct{})
			cancel := func() {
				ctx.Info("[SERVER] Canceling downlink")
				close(stop)
			}

			go func() {
				ctx.Info("[SERVER] Uplink channel opened")
				for message := range up {
					ctx.WithField("Message", message).Info("[SERVER] Received Uplink")
				}
				ctx.Info("[SERVER] Uplink channel closed")
			}()

			go func() {
			loop:
				for {
					select {
					case <-stop:
						break loop
					case <-time.After(5 * time.Millisecond):
						ctx.Info("[SERVER] Sending Downlink")
						down <- &DownlinkMessage{
							Payload: []byte{1, 2, 3, 4},
						}
					}
				}
				close(down)
				ctx.Info("[SERVER] Closed Downlink")
			}()

			return up, down, cancel, nil
		}

		brkClient := NewBrokerClient(conn)
		stream := NewMonitoredRouterStream(brkClient, func() context.Context {
			return context.Background()
		})

		ch := stream.Channel()

		go func() {
			for downlink := range ch {
				ctx.WithField("Downlink", downlink).Info("[CLIENT] Received Downlink")
			}
			ctx.Info("[CLIENT] Closed Downlink")
		}()

		err := stream.Send(&UplinkMessage{
			Payload: []byte{1, 2, 3, 4},
		})

		a.So(err, ShouldBeNil)

		time.Sleep(10 * time.Millisecond)

		stream.Close()

		time.Sleep(10 * time.Millisecond)
	}

}
