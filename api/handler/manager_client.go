package handler

import (
	"encoding/json"
	"os"
	"os/user"
	"sync"

	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

func (h *ManagerClient) getContext() context.Context {
	h.RLock()
	defer h.RUnlock()
	md := metadata.Pairs(
		"id", h.id,
		"token", h.accessToken,
	)
	return metadata.NewContext(context.Background(), md)
}

// GetApplication retrieves an application from the Handler
func (h *ManagerClient) GetApplication(appID string) (*Application, error) {
	res, err := h.applicationManagerClient.GetApplication(h.getContext(), &ApplicationIdentifier{AppId: appID})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not get application from Handler")
	}
	return res, nil
}

// SetApplication sets an application on the Handler
func (h *ManagerClient) SetApplication(in *Application) error {
	_, err := h.applicationManagerClient.SetApplication(h.getContext(), in)
	return errors.Wrap(errors.FromGRPCError(err), "Could not set application on Handler")
}

// RegisterApplication registers an application on the Handler
func (h *ManagerClient) RegisterApplication(appID string) error {
	_, err := h.applicationManagerClient.RegisterApplication(h.getContext(), &ApplicationIdentifier{AppId: appID})
	return errors.Wrap(errors.FromGRPCError(err), "Could not register application on Handler")
}

// DeleteApplication deletes an application and all its devices from the Handler
func (h *ManagerClient) DeleteApplication(appID string) error {
	_, err := h.applicationManagerClient.DeleteApplication(h.getContext(), &ApplicationIdentifier{AppId: appID})
	return errors.Wrap(errors.FromGRPCError(err), "Could not delete application from Handler")
}

// GetDevice retrieves a device from the Handler
func (h *ManagerClient) GetDevice(appID string, devID string) (*Device, error) {
	res, err := h.applicationManagerClient.GetDevice(h.getContext(), &DeviceIdentifier{AppId: appID, DevId: devID})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not get device from Handler")
	}
	return res, nil
}

// SetDevice sets a device on the Handler
func (h *ManagerClient) SetDevice(in *Device) error {
	_, err := h.applicationManagerClient.SetDevice(h.getContext(), in)
	return errors.Wrap(errors.FromGRPCError(err), "Could not set device on Handler")
}

// DeleteDevice deletes a device from the Handler
func (h *ManagerClient) DeleteDevice(appID string, devID string) error {
	_, err := h.applicationManagerClient.DeleteDevice(h.getContext(), &DeviceIdentifier{AppId: appID, DevId: devID})
	return errors.Wrap(errors.FromGRPCError(err), "Could not delete device from Handler")
}

// GetDevicesForApplication retrieves all devices for an application from the Handler
func (h *ManagerClient) GetDevicesForApplication(appID string) (devices []*Device, err error) {
	res, err := h.applicationManagerClient.GetDevicesForApplication(h.getContext(), &ApplicationIdentifier{AppId: appID})
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
	resp, err := devAddrManager.GetDevAddr(h.getContext(), &lorawan.DevAddrRequest{
		Usage: constraints,
	})
	if err != nil {
		return types.DevAddr{}, errors.Wrap(errors.FromGRPCError(err), "Could not get DevAddr from Handler")
	}
	return *resp.DevAddr, nil
}

// DryUplink transforms the uplink payload with the payload functions provided
// in the app..
func (h *ManagerClient) DryUplink(payload []byte, app *Application) (*DryUplinkResult, error) {
	res, err := h.applicationManagerClient.DryUplink(h.getContext(), &DryUplinkMessage{
		App:     app,
		Payload: payload,
	})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not dry-run uplink on Handler")
	}
	return res, nil
}

// DryDownlinkWithPayload transforms the downlink payload with the payload functions
// provided in app.
func (h *ManagerClient) DryDownlinkWithPayload(payload []byte, app *Application) (*DryDownlinkResult, error) {
	res, err := h.applicationManagerClient.DryDownlink(h.getContext(), &DryDownlinkMessage{
		App:     app,
		Payload: payload,
	})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not dry-run downlink with payload on Handler")
	}
	return res, nil
}

// DryDownlinkWithFields transforms the downlink fields with the payload functions
// provided in app.
func (h *ManagerClient) DryDownlinkWithFields(fields map[string]interface{}, app *Application) (*DryDownlinkResult, error) {
	marshalled, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}

	res, err := h.applicationManagerClient.DryDownlink(h.getContext(), &DryDownlinkMessage{
		App:    app,
		Fields: string(marshalled),
	})
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "Could not dry-run downlink with fields on Handler")
	}
	return res, nil
}

// Close closes the client
func (h *ManagerClient) Close() error {
	return h.conn.Close()
}
