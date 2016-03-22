// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type brokerClient struct {
	sync.Mutex
	*tokenCredentials
	core.BrokerClient
	core.BrokerManagerClient
}

// NewClient instantiates a new core.Broker client
func NewClient(netAddr string) (core.Broker, error) {
	brokerConn, err := grpc.Dial(netAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second*15)) // Add Credentials Token
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

// BeginToken implements the core.Broker interface
func (b *brokerClient) BeginToken(token string) core.Broker {
	b.Lock()
	b.token = token
	return b
}

// EndToken implements the core.Broker interface
func (b *brokerClient) EndToken() {
	b.Unlock()
}

type tokenCredentials struct {
	token string
}

// GetRequestMetadata implements the grpc/credentials.Credentials interface
func (t tokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"token": t.token,
	}, nil
}

// RequireTransportSecurity implements the grpc/credentials.Credentials interface
func (t tokenCredentials) RequireTransportSecurity() bool {
	return false // TODO -> True, Need use of TLS to communicate the token
}
