// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_handler "github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/api/monitor/monitorclient"
	pb "github.com/TheThingsNetwork/api/networkserver"
	"github.com/TheThingsNetwork/go-utils/grpc/auth"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"google.golang.org/grpc"
	"gopkg.in/redis.v5"
)

// NetworkServer implements LoRaWAN-specific functionality for TTN
type NetworkServer interface {
	component.Interface
	component.ManagementInterface

	UsePrefix(prefix types.DevAddrPrefix, usage []string) error
	GetPrefixesFor(requiredUsages ...string) []types.DevAddrPrefix

	HandleGetDevices(*pb.DevicesRequest) (*pb.DevicesResponse, error)
	HandlePrepareActivation(*pb_broker.DeduplicatedDeviceActivationRequest) (*pb_broker.DeduplicatedDeviceActivationRequest, error)
	HandleActivate(*pb_handler.DeviceActivationResponse) (*pb_handler.DeviceActivationResponse, error)
	HandleUplink(*pb_broker.DeduplicatedUplinkMessage) (*pb_broker.DeduplicatedUplinkMessage, error)
	HandleDownlink(*pb_broker.DownlinkMessage) (*pb_broker.DownlinkMessage, error)
}

// NewRedisNetworkServer creates a new Redis-backed NetworkServer
func NewRedisNetworkServer(client *redis.Client, netID int) NetworkServer {
	ns := &networkServer{
		devices:  device.NewRedisDeviceStore(client, "ns"),
		prefixes: map[types.DevAddrPrefix][]string{},
	}
	ns.netID = [3]byte{byte(netID >> 16), byte(netID >> 8), byte(netID)}
	return ns
}

type networkServer struct {
	*component.Component
	devices       device.Store
	netID         [3]byte
	prefixes      map[types.DevAddrPrefix][]string
	status        *status
	monitorStream monitorclient.Stream
}

func (n *networkServer) UsePrefix(prefix types.DevAddrPrefix, usage []string) error {
	if prefix.Length < 7 {
		return errors.NewErrInvalidArgument("Prefix", "invalid length")
	}
	if prefix.DevAddr[0]>>1 != n.netID[2] {
		return errors.NewErrInvalidArgument("Prefix", "invalid netID")
	}
	n.prefixes[prefix] = usage
	return nil
}

func (n *networkServer) GetPrefixesFor(requiredUsages ...string) []types.DevAddrPrefix {
	var suitablePrefixes []types.DevAddrPrefix
	for prefix, offeredUsages := range n.prefixes {
		matches := 0
		for _, requiredUsage := range requiredUsages {
			for _, offeredUsage := range offeredUsages {
				if offeredUsage == requiredUsage {
					matches++
				}
			}
		}
		if matches == len(requiredUsages) {
			suitablePrefixes = append(suitablePrefixes, prefix)
		}
	}
	return suitablePrefixes
}

func (n *networkServer) Init(c *component.Component) error {
	n.Component = c
	n.InitStatus()
	err := n.Component.UpdateTokenKey()
	if err != nil {
		return err
	}
	n.Component.SetStatus(component.StatusHealthy)
	if n.Component.Monitor != nil {
		n.monitorStream = n.Component.Monitor.NetworkServerClient(n.Context, grpc.PerRPCCredentials(auth.WithStaticToken(n.AccessToken)))
	}
	return nil
}

func (n *networkServer) Shutdown() {}
