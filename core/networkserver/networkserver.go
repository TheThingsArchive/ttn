// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	pb "github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"gopkg.in/redis.v4"
)

// NetworkServer implements LoRaWAN-specific functionality for TTN
type NetworkServer interface {
	core.ComponentInterface
	core.ManagementInterface

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
	*core.Component
	devices  device.Store
	netID    [3]byte
	prefixes map[types.DevAddrPrefix][]string
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

func (n *networkServer) Init(c *core.Component) error {
	n.Component = c
	err := n.Component.UpdateTokenKey()
	if err != nil {
		return err
	}
	n.Component.SetStatus(core.StatusHealthy)
	return nil
}

func (n *networkServer) Shutdown() {}
