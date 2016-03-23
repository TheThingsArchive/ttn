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

// AppClient mocks the core.AppClient interface
type AppClient struct {
	Failures     map[string]error
	InHandleData struct {
		Ctx  context.Context
		Req  *core.DataAppReq
		Opts []grpc.CallOption
	}
	OutHandleData struct {
		Res *core.DataAppRes
	}
	InHandleJoin struct {
		Ctx  context.Context
		Req  *core.JoinAppReq
		Opts []grpc.CallOption
	}
	OutHandleJoin struct {
		Res *core.JoinAppRes
	}
}

// NewAppClient creates a new mock AppClient
func NewAppClient() *AppClient {
	return &AppClient{
		Failures: make(map[string]error),
	}
}

// HandleJoin implements the core.AppClient interface
func (m *AppClient) HandleJoin(ctx context.Context, in *core.JoinAppReq, opts ...grpc.CallOption) (*core.JoinAppRes, error) {
	m.InHandleJoin.Ctx = ctx
	m.InHandleJoin.Req = in
	m.InHandleJoin.Opts = opts
	return m.OutHandleJoin.Res, m.Failures["HandleJoin"]
}

// HandleData implements the core.AppClient interface
func (m *AppClient) HandleData(ctx context.Context, in *core.DataAppReq, opts ...grpc.CallOption) (*core.DataAppRes, error) {
	m.InHandleData.Ctx = ctx
	m.InHandleData.Req = in
	m.InHandleData.Opts = opts
	return m.OutHandleData.Res, m.Failures["HandleData"]
}

// HandlerClient mocks the core.HandlerClient interface
type HandlerClient struct {
	Failures     map[string]error
	InHandleJoin struct {
		Ctx  context.Context
		Req  *core.JoinHandlerReq
		Opts []grpc.CallOption
	}
	OutHandleJoin struct {
		Res *core.JoinHandlerRes
	}
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
}

// NewHandlerClient creates a new mock HandlerClient
func NewHandlerClient() *HandlerClient {
	return &HandlerClient{
		Failures: make(map[string]error),
	}
}

// HandleJoin implements the core.HandlerClient interface
func (m *HandlerClient) HandleJoin(ctx context.Context, in *core.JoinHandlerReq, opts ...grpc.CallOption) (*core.JoinHandlerRes, error) {
	m.InHandleJoin.Ctx = ctx
	m.InHandleJoin.Req = in
	m.InHandleJoin.Opts = opts
	return m.OutHandleJoin.Res, m.Failures["HandleJoin"]
}

// HandleDataUp implements the core.HandlerClient interface
func (m *HandlerClient) HandleDataUp(ctx context.Context, in *core.DataUpHandlerReq, opts ...grpc.CallOption) (*core.DataUpHandlerRes, error) {
	m.InHandleDataUp.Ctx = ctx
	m.InHandleDataUp.Req = in
	m.InHandleDataUp.Opts = opts
	return m.OutHandleDataUp.Res, m.Failures["HandleDataUp"]
}

// HandleDataDown implements the core.HandlerClient interface
func (m *HandlerClient) HandleDataDown(ctx context.Context, in *core.DataDownHandlerReq, opts ...grpc.CallOption) (*core.DataDownHandlerRes, error) {
	m.InHandleDataDown.Ctx = ctx
	m.InHandleDataDown.Req = in
	m.InHandleDataDown.Opts = opts
	return m.OutHandleDataDown.Res, m.Failures["HandleDataDown"]
}

// AuthBrokerClient mocks the core.AuthBrokerClient interface
type AuthBrokerClient struct {
	Failures     map[string]error
	InHandleData struct {
		Ctx  context.Context
		Req  *core.DataBrokerReq
		Opts []grpc.CallOption
	}
	OutHandleData struct {
		Res *core.DataBrokerRes
	}
	InHandleJoin struct {
		Ctx  context.Context
		Req  *core.JoinBrokerReq
		Opts []grpc.CallOption
	}
	OutHandleJoin struct {
		Res *core.JoinBrokerRes
	}
	InUpsertABP struct {
		Ctx  context.Context
		Req  *core.UpsertABPBrokerReq
		Opts []grpc.CallOption
	}
	OutUpsertABP struct {
		Res *core.UpsertABPBrokerRes
	}
	InValidateOTAA struct {
		Ctx  context.Context
		Req  *core.ValidateOTAABrokerReq
		Opts []grpc.CallOption
	}
	OutValidateOTAA struct {
		Res *core.ValidateOTAABrokerRes
	}
	InValidateToken struct {
		Ctx  context.Context
		Req  *core.ValidateTokenBrokerReq
		Opts []grpc.CallOption
	}
	OutValidateToken struct {
		Res *core.ValidateTokenBrokerRes
	}
}

// NewAuthBrokerClient creates a new mock AuthBrokerClient
func NewAuthBrokerClient() *AuthBrokerClient {
	return &AuthBrokerClient{
		Failures: make(map[string]error),
	}
}

// HandleData implements the core.Broker interface
func (m *AuthBrokerClient) HandleData(ctx context.Context, in *core.DataBrokerReq, opts ...grpc.CallOption) (*core.DataBrokerRes, error) {
	m.InHandleData.Ctx = ctx
	m.InHandleData.Req = in
	m.InHandleData.Opts = opts
	return m.OutHandleData.Res, m.Failures["HandleData"]
}

// HandleJoin implements the core.Broker interface
func (m *AuthBrokerClient) HandleJoin(ctx context.Context, in *core.JoinBrokerReq, opts ...grpc.CallOption) (*core.JoinBrokerRes, error) {
	m.InHandleJoin.Ctx = ctx
	m.InHandleJoin.Req = in
	m.InHandleJoin.Opts = opts
	return m.OutHandleJoin.Res, m.Failures["HandleJoin"]
}

// UpsertABP implements the core.Broker interface
func (m *AuthBrokerClient) UpsertABP(ctx context.Context, in *core.UpsertABPBrokerReq, opts ...grpc.CallOption) (*core.UpsertABPBrokerRes, error) {
	m.InUpsertABP.Ctx = ctx
	m.InUpsertABP.Req = in
	m.InUpsertABP.Opts = opts
	return m.OutUpsertABP.Res, m.Failures["UpsertABP"]
}

// ValidateOTAA implements the core.Broker interface
func (m *AuthBrokerClient) ValidateOTAA(ctx context.Context, in *core.ValidateOTAABrokerReq, opts ...grpc.CallOption) (*core.ValidateOTAABrokerRes, error) {
	m.InValidateOTAA.Ctx = ctx
	m.InValidateOTAA.Req = in
	m.InValidateOTAA.Opts = opts
	return m.OutValidateOTAA.Res, m.Failures["ValidateOTAA"]
}

// ValidateToken implements the core.Broker interface
func (m *AuthBrokerClient) ValidateToken(ctx context.Context, in *core.ValidateTokenBrokerReq, opts ...grpc.CallOption) (*core.ValidateTokenBrokerRes, error) {
	m.InValidateToken.Ctx = ctx
	m.InValidateToken.Req = in
	m.InValidateToken.Opts = opts
	return m.OutValidateToken.Res, m.Failures["ValidateToken"]
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
	InHandleJoin struct {
		Ctx context.Context
		Req *core.JoinRouterReq
	}
	OutHandleJoin struct {
		Res *core.JoinRouterRes
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
	return m.OutHandleData.Res, m.Failures["HandleData"]
}

// HandleStats implements the core.RouterServer interface
func (m *RouterServer) HandleStats(ctx context.Context, in *core.StatsReq) (*core.StatsRes, error) {
	m.InHandleStats.Ctx = ctx
	m.InHandleStats.Req = in
	return m.OutHandleStats.Res, m.Failures["HandleStats"]
}

// HandleJoin implements the core.RouterServer interface
func (m *RouterServer) HandleJoin(ctx context.Context, in *core.JoinRouterReq) (*core.JoinRouterRes, error) {
	m.InHandleJoin.Ctx = ctx
	m.InHandleJoin.Req = in
	return m.OutHandleJoin.Res, m.Failures["HandleJoin"]
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

// HandlerServer mocks the core.HandlerServer interface
type HandlerServer struct {
	Failures       map[string]error
	InHandleDataUp struct {
		Ctx context.Context
		Req *core.DataUpHandlerReq
	}
	OutHandleDataUp struct {
		Res *core.DataUpHandlerRes
	}
	InHandleDataDown struct {
		Ctx context.Context
		Req *core.DataDownHandlerReq
	}
	OutHandleDataDown struct {
		Res *core.DataDownHandlerRes
	}
	InHandleJoin struct {
		Ctx context.Context
		Req *core.JoinHandlerReq
	}
	OutHandleJoin struct {
		Res *core.JoinHandlerRes
	}
}

// NewHandlerServer creates a new mock HandlerServer
func NewHandlerServer() *HandlerServer {
	return &HandlerServer{
		Failures: make(map[string]error),
	}
}

// HandleDataUp implements the core.HandlerServer interface
func (m *HandlerServer) HandleDataUp(ctx context.Context, in *core.DataUpHandlerReq) (*core.DataUpHandlerRes, error) {
	m.InHandleDataUp.Ctx = ctx
	m.InHandleDataUp.Req = in
	return m.OutHandleDataUp.Res, m.Failures["HandleDataUp"]
}

// HandleDataDown implements the core.HandlerServer interface
func (m *HandlerServer) HandleDataDown(ctx context.Context, in *core.DataDownHandlerReq) (*core.DataDownHandlerRes, error) {
	m.InHandleDataDown.Ctx = ctx
	m.InHandleDataDown.Req = in
	return m.OutHandleDataDown.Res, m.Failures["HandleDataDown"]
}

// HandleJoin implements the core.HandlerServer interface
func (m *HandlerServer) HandleJoin(ctx context.Context, in *core.JoinHandlerReq) (*core.JoinHandlerRes, error) {
	m.InHandleJoin.Ctx = ctx
	m.InHandleJoin.Req = in
	return m.OutHandleJoin.Res, m.Failures["HandleJoin"]
}
