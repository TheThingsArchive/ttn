// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type handlerClient struct {
	sync.Mutex
	*tokenCredentials
	core.HandlerClient
	core.HandlerManagerClient
}

// NewClient instantiates a new core.handler client
func NewClient(netAddr string) (core.AuthHandlerClient, error) {
	tokener := tokenCredentials{}
	handlerConn, err := grpc.Dial(
		netAddr,
		grpc.WithInsecure(), // TODO Use of TLS
		grpc.WithPerRPCCredentials(&tokener),
		grpc.WithTimeout(time.Second*15),
	)
	if err != nil {
		return nil, err
	}
	handler := core.NewHandlerClient(handlerConn)
	handlerManager := core.NewHandlerManagerClient(handlerConn)
	return &handlerClient{
		tokenCredentials:     &tokener,
		HandlerClient:        handler,
		HandlerManagerClient: handlerManager,
	}, nil
}

// BeginToken implements the core.handler interface
func (b *handlerClient) BeginToken(token string) core.AuthHandlerClient {
	b.Lock()
	b.token = token
	return b
}

// EndToken implements the core.handler interface
func (b *handlerClient) EndToken() {
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
