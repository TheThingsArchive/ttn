// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package mocks offers dedicated mocking interface / structures for testing
package mocks

import (
	"github.com/TheThingsNetwork/ttn/core"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NOTE: All the code below could be generated

// HandlerClient mocks the core.HandlerClient interface
type HandlerClient struct {
	Failures       map[string]error
	InHandleDataUp struct {
		Ctx  context.Context
		Req  *core.DataUpHandlerReq
		Opts []grpc.CallOption
	}
	OutHandleDataUp struct {
		Res *core.DataUpHandlerRes
	}
	InHandleDataDown struct {
		Ctx  context.Context
		Req  *core.DataDownHandlerReq
		Opts []grpc.CallOption
	}
	OutHandleDataDown struct {
		Res *core.DataDownHandlerRes
	}
	InSubscribePersonalized struct {
		Ctx  context.Context
		Req  *core.ABPSubHandlerReq
		Opts []grpc.CallOption
	}
	OutSubscribePersonalized struct {
		Res *core.ABPSubHandlerRes
	}
}

// NewHandlerClient creates a new mock HandlerClient
func NewHandlerClient() *HandlerClient {
	return &HandlerClient{
		Failures: make(map[string]error),
	}
}

// HandleDataUp implements the core.HandlerClient interface
func (m *HandlerClient) HandleDataUp(ctx context.Context, in *core.DataUpHandlerReq, opts ...grpc.CallOption) (*core.DataUpHandlerRes, error) {
	m.InHandleDataUp.Ctx = ctx
	m.InHandleDataUp.Req = in
	m.InHandleDataUp.Opts = opts
	if err := m.Failures["HandleDataUp"]; err != nil {
		return nil, err
	}
	return m.OutHandleDataUp.Res, nil
}

// HandleDataDown implements the core.HandlerClient interface
func (m *HandlerClient) HandleDataDown(ctx context.Context, in *core.DataDownHandlerReq, opts ...grpc.CallOption) (*core.DataDownHandlerRes, error) {
	m.InHandleDataDown.Ctx = ctx
	m.InHandleDataDown.Req = in
	m.InHandleDataDown.Opts = opts
	if err := m.Failures["HandleDataDown"]; err != nil {
		return nil, err
	}
	return m.OutHandleDataDown.Res, nil
}

// SubscribePersonalized implements the core.HandlerClient interface
func (m *HandlerClient) SubscribePersonalized(ctx context.Context, in *core.ABPSubHandlerReq, opts ...grpc.CallOption) (*core.ABPSubHandlerRes, error) {
	m.InSubscribePersonalized.Ctx = ctx
	m.InSubscribePersonalized.Req = in
	m.InSubscribePersonalized.Opts = opts
	if err := m.Failures["SubscribePersonalized"]; err != nil {
		return nil, err
	}
	return m.OutSubscribePersonalized.Res, nil
}
