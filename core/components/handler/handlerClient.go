// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"google.golang.org/grpc"
)

type handlerClient struct {
	core.HandlerClient
	core.HandlerManagerClient
}

// NewClient instantiates a new core.handler client
func NewClient(netAddr string) (core.AuthHandlerClient, error) {
	handlerConn, err := grpc.Dial(
		netAddr,
		grpc.WithInsecure(), // TODO Use of TLS
		grpc.WithTimeout(time.Second*15),
	)
	if err != nil {
		return nil, err
	}
	handler := core.NewHandlerClient(handlerConn)
	handlerManager := core.NewHandlerManagerClient(handlerConn)
	return &handlerClient{
		HandlerClient:        handler,
		HandlerManagerClient: handlerManager,
	}, nil
}
