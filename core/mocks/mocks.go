// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package mocks offers dedicated mocking interface / structures for testing
package mocks

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/dutycycle"
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

// BrokerClient mocks the core.BrokerClient interface
type BrokerClient struct {
	Failures     map[string]error
	InHandleData struct {
		Ctx  context.Context
		Req  *core.DataBrokerReq
		Opts []grpc.CallOption
	}
	OutHandleData struct {
		Res *core.DataBrokerRes
	}
	InSubscribePersonalized struct {
		Ctx  context.Context
		Req  *core.ABPSubBrokerReq
		Opts []grpc.CallOption
	}
	OutSubscribePersonalized struct {
		Res *core.ABPSubBrokerRes
	}
}

// NewBrokerClient creates a new mock BrokerClient
func NewBrokerClient() *BrokerClient {
	return &BrokerClient{
		Failures: make(map[string]error),
	}
}

// HandleData implements the core.BrokerClient interface
func (m *BrokerClient) HandleData(ctx context.Context, in *core.DataBrokerReq, opts ...grpc.CallOption) (*core.DataBrokerRes, error) {
	m.InHandleData.Ctx = ctx
	m.InHandleData.Req = in
	m.InHandleData.Opts = opts
	if err := m.Failures["HandleData"]; err != nil {
		return nil, err
	}
	return m.OutHandleData.Res, nil
}

// SubscribePersonalized implements the core.BrokerClient interface
func (m *BrokerClient) SubscribePersonalized(ctx context.Context, in *core.ABPSubBrokerReq, opts ...grpc.CallOption) (*core.ABPSubBrokerRes, error) {
	m.InSubscribePersonalized.Ctx = ctx
	m.InSubscribePersonalized.Req = in
	m.InSubscribePersonalized.Opts = opts
	if err := m.Failures["SubscribePersonalized"]; err != nil {
		return nil, err
	}
	return m.OutSubscribePersonalized.Res, nil
}

// RouterServer mocks the core.RouterServer interface
type RouterServer struct {
	Failures     map[string]error
	InHandleData struct {
		Ctx context.Context
		Req *core.DataRouterReq
	}
	OutHandleData struct {
		Res *core.DataRouterRes
	}
	InHandleStats struct {
		Ctx context.Context
		Req *core.StatsReq
	}
	OutHandleStats struct {
		Res *core.StatsRes
	}
}

// NewRouterServer creates a new mock RouterServer
func NewRouterServer() *RouterServer {
	return &RouterServer{
		Failures: make(map[string]error),
	}
}

// HandleData implements the core.RouterServer interface
func (m *RouterServer) HandleData(ctx context.Context, in *core.DataRouterReq) (*core.DataRouterRes, error) {
	m.InHandleData.Ctx = ctx
	m.InHandleData.Req = in
	if err := m.Failures["HandleData"]; err != nil {
		return nil, err
	}
	return m.OutHandleData.Res, nil
}

// HandleStats implements the core.RouterServer interface
func (m *RouterServer) HandleStats(ctx context.Context, in *core.StatsReq) (*core.StatsRes, error) {
	m.InHandleStats.Ctx = ctx
	m.InHandleStats.Req = in
	if err := m.Failures["HandleStats"]; err != nil {
		return nil, err
	}
	return m.OutHandleStats.Res, nil
}

// DutyManager mocks the dutycycle.DutyManager interface
type DutyManager struct {
	Failures map[string]error
	InUpdate struct {
		ID   []byte
		Freq float32
		Size uint32
		Datr string
		Codr string
	}
	InLookup struct {
		ID []byte
	}
	OutLookup struct {
		Cycles dutycycle.Cycles
	}
	InClose struct {
		Called bool
	}
}

// NewDutyManager creates a new mock DutyManager
func NewDutyManager() *DutyManager {
	return &DutyManager{
		Failures: make(map[string]error),
	}
}

// Update implements the dutycycle.DutyManager interface
func (m *DutyManager) Update(id []byte, freq float32, size uint32, datr string, codr string) error {
	m.InUpdate.ID = id
	m.InUpdate.Freq = freq
	m.InUpdate.Size = size
	m.InUpdate.Datr = datr
	m.InUpdate.Codr = codr
	return m.Failures["Update"]
}

// Lookup implements the dutycycle.DutyManager interface
func (m *DutyManager) Lookup(id []byte) (dutycycle.Cycles, error) {
	m.InLookup.ID = id
	return m.OutLookup.Cycles, m.Failures["Lookup"]
}

// Close implements the dutycycle.DutyManager interface
func (m *DutyManager) Close() error {
	m.InClose.Called = true
	return m.Failures["Close"]
}
