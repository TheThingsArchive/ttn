// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"google.golang.org/grpc"
)

type brokerClient struct {
	core.BrokerClient
	core.BrokerManagerClient
}

// NewClient instantiates a new core.Broker client
func NewClient(netAddr string) (core.AuthBrokerClient, error) {
	brokerConn, err := grpc.Dial(
		netAddr,
		grpc.WithInsecure(), // TODO Use of TLS
		grpc.WithTimeout(time.Second*15),
	)
	if err != nil {
		return nil, err
	}
	broker := core.NewBrokerClient(brokerConn)
	brokerManager := core.NewBrokerManagerClient(brokerConn)
	return &brokerClient{
		BrokerClient:        broker,
		BrokerManagerClient: brokerManager,
	}, nil
}
