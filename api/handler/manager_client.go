// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"encoding/json"
	"os"
	"os/user"
	"sync"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
)

// ManagerClient is used to manage applications and devices on a handler
type ManagerClient struct {
	sync.RWMutex
	id                       string
	accessToken              string
	conn                     *grpc.ClientConn
	applicationManagerClient ApplicationManagerClient
}

// NewManagerClient returns a new ManagerClient for a handler on the given conn that accepts the given access token
func NewManagerClient(conn *grpc.ClientConn, accessToken string) (*ManagerClient, error) {
	applicationManagerClient := NewApplicationManagerClient(conn)

	id := "client"
	if user, err := user.Current(); err == nil {
		id += "-" + user.Username
	}
	if hostname, err := os.Hostname(); err == nil {
		id += "@" + hostname
	}

	return &ManagerClient{
		id:          id,
		accessToken: accessToken,
		conn:        conn,
		applicationManagerClient: applicationManagerClient,
	}, nil
}

// SetID sets the ID of this client
func (h *ManagerClient) SetID(id string) {
	h.Lock()
	defer h.Unlock()
	h.id = id
}

// UpdateAccessToken updates the access token that is used for running commands
func (h *ManagerClient) UpdateAccessToken(accessToken string) {
	h.Lock()
	defer h.Unlock()
	h.accessToken = accessToken
}

// GetContext returns a new context with authentication
func (h *ManagerClient) GetContext() context.Context {
	h.RLock()
	defer h.RUnlock()
	ctx := context.Background()
	ctx = api.ContextWithID(ctx, h.id)
	ctx = api.ContextWithToken(ctx, h.accessToken)
	return ctx
}

// GetContext returns a new context with authentication, plus limit and offset for pagination
func (h *ManagerClient) GetContextWithLimitAndOffset(limit, offset int) context.Context {
	h.RLock()
	defer h.RUnlock()
	ctx := h.GetContext()
	ctx = api.ContextWithLimitAndOffset(ctx, uint64(limit), uint64(offset))
	return ctx
}

// GetApplication retrieves an application from the Handler
func (h *ManagerClient) GetApplication(appID string) (*Application, error) {
	res, err := h.applicationManagerClient.GetApplication(h.GetContext(), &ApplicationIdentifier{AppId: appID})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not get application from Handler")
	}
	return res, nil
}

// SetApplication sets an application on the Handler
func (h *ManagerClient) SetApplication(in *Application) error {
	_, err := h.applicationManagerClient.SetApplication(h.GetContext(), in)
	return errors.Wrap(errors.FromGRPCError(err), "Could not set application on Handler")
}

// RegisterApplication registers an application on the Handler
func (h *ManagerClient) RegisterApplication(appID string) error {
	_, err := h.applicationManagerClient.RegisterApplication(h.GetContext(), &ApplicationIdentifier{AppId: appID})
	return errors.Wrap(errors.FromGRPCError(err), "Could not register application on Handler")
}

// DeleteApplication deletes an application and all its devices from the Handler
func (h *ManagerClient) DeleteApplication(appID string) error {
	_, err := h.applicationManagerClient.DeleteApplication(h.GetContext(), &ApplicationIdentifier{AppId: appID})
	return errors.Wrap(errors.FromGRPCError(err), "Could not delete application from Handler")
}

// GetDevice retrieves a device from the Handler
func (h *ManagerClient) GetDevice(appID string, devID string) (*Device, error) {
	res, err := h.applicationManagerClient.GetDevice(h.GetContext(), &DeviceIdentifier{AppId: appID, DevId: devID})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not get device from Handler")
	}
	return res, nil
}

// SetDevice sets a device on the Handler
func (h *ManagerClient) SetDevice(in *Device) error {
	_, err := h.applicationManagerClient.SetDevice(h.GetContext(), in)
	return errors.Wrap(errors.FromGRPCError(err), "Could not set device on Handler")
}

// DeleteDevice deletes a device from the Handler
func (h *ManagerClient) DeleteDevice(appID string, devID string) error {
	_, err := h.applicationManagerClient.DeleteDevice(h.GetContext(), &DeviceIdentifier{AppId: appID, DevId: devID})
	return errors.Wrap(errors.FromGRPCError(err), "Could not delete device from Handler")
}

// GetDevicesForApplication retrieves all devices for an application from the Handler.
// Pass a limit to indicate the maximum number of results you want to receive, and the offset to indicate how many results should be skipped.
func (h *ManagerClient) GetDevicesForApplication(appID string, limit, offset int) (devices []*Device, err error) {
	res, err := h.applicationManagerClient.GetDevicesForApplication(h.GetContextWithLimitAndOffset(limit, offset), &ApplicationIdentifier{AppId: appID})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not get devices for application from Handler")
	}
	for _, dev := range res.Devices {
		devices = append(devices, dev)
	}
	return
}

// GetDevAddr requests a random device address with the given constraints
func (h *ManagerClient) GetDevAddr(constraints ...string) (types.DevAddr, error) {
	devAddrManager := lorawan.NewDevAddrManagerClient(h.conn)
	resp, err := devAddrManager.GetDevAddr(h.GetContext(), &lorawan.DevAddrRequest{
		Usage: constraints,
	})
	if err != nil {
		return types.DevAddr{}, errors.Wrap(errors.FromGRPCError(err), "Could not get DevAddr from Handler")
	}
	return *resp.DevAddr, nil
}

// DryUplink transforms the uplink payload with the payload functions provided
// in the app..
func (h *ManagerClient) DryUplink(payload []byte, app *Application, port uint32) (*DryUplinkResult, error) {
	res, err := h.applicationManagerClient.DryUplink(h.GetContext(), &DryUplinkMessage{
		App:     app,
		Payload: payload,
		Port:    port,
	})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not dry-run uplink on Handler")
	}
	return res, nil
}

// DryDownlinkWithPayload transforms the downlink payload with the payload functions
// provided in app.
func (h *ManagerClient) DryDownlinkWithPayload(payload []byte, app *Application, port uint32) (*DryDownlinkResult, error) {
	res, err := h.applicationManagerClient.DryDownlink(h.GetContext(), &DryDownlinkMessage{
		App:     app,
		Payload: payload,
		Port:    port,
	})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not dry-run downlink with payload on Handler")
	}
	return res, nil
}

// DryDownlinkWithFields transforms the downlink fields with the payload functions
// provided in app.
func (h *ManagerClient) DryDownlinkWithFields(fields map[string]interface{}, app *Application, port uint32) (*DryDownlinkResult, error) {
	marshalled, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}

	res, err := h.applicationManagerClient.DryDownlink(h.GetContext(), &DryDownlinkMessage{
		App:    app,
		Fields: string(marshalled),
		Port:   port,
	})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not dry-run downlink with fields on Handler")
	}
	return res, nil
}

// SimulateUplink simulates an uplink message
func (h *ManagerClient) SimulateUplink(appID string, devID string, port uint32, payload []byte) error {
	_, err := h.applicationManagerClient.SimulateUplink(h.GetContext(), &SimulatedUplinkMessage{
		AppId:   appID,
		DevId:   devID,
		Port:    port,
		Payload: payload,
	})
	if err != nil {
		return errors.Wrap(errors.FromGRPCError(err), "Could not simulate uplink")
	}
	return nil
}

// Close closes the client
func (h *ManagerClient) Close() error {
	return h.conn.Close()
}
