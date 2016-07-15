package handler

import (
	"fmt"
	"sync"

	"github.com/TheThingsNetwork/ttn/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ManagerClient is used to manage applications and devices on a handler
type ManagerClient struct {
	sync.RWMutex
	conn                     *grpc.ClientConn
	context                  context.Context
	applicationManagerClient ApplicationManagerClient
}

// NewManagerClient returns a new ManagerClient for a handler at the given address that accepts the given access token
func NewManagerClient(address string, accessToken string) (*ManagerClient, error) {
	conn, err := grpc.Dial(address, api.DialOptions...)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to NetworkServer: %s", err.Error())
	}
	applicationManagerClient := NewApplicationManagerClient(conn)
	md := metadata.Pairs(
		"token", accessToken,
	)
	manageContext := metadata.NewContext(context.Background(), md)
	return &ManagerClient{
		conn:                     conn,
		context:                  manageContext,
		applicationManagerClient: applicationManagerClient,
	}, nil
}

// UpdateAccessToken updates the access token that is used for running commands
func (h *ManagerClient) UpdateAccessToken(accessToken string) {
	h.Lock()
	defer h.Unlock()
	md := metadata.Pairs(
		"token", accessToken,
	)
	h.context = metadata.NewContext(context.Background(), md)
}

// GetApplication retrieves an application from the Handler
func (h *ManagerClient) GetApplication(appID string) (*Application, error) {
	h.RLock()
	defer h.RUnlock()
	return h.applicationManagerClient.GetApplication(h.context, &ApplicationIdentifier{AppId: appID})
}

// SetApplication sets an application on the Handler
func (h *ManagerClient) SetApplication(in *Application) error {
	h.RLock()
	defer h.RUnlock()
	_, err := h.applicationManagerClient.SetApplication(h.context, in)
	return err
}

// DeleteApplication deletes an application and all its devices from the Handler
func (h *ManagerClient) DeleteApplication(appID string) error {
	h.RLock()
	defer h.RUnlock()
	_, err := h.applicationManagerClient.DeleteApplication(h.context, &ApplicationIdentifier{AppId: appID})
	return err
}

// GetDevice retrieves a device from the Handler
func (h *ManagerClient) GetDevice(appID string, devID string) (*Device, error) {
	h.RLock()
	defer h.RUnlock()
	return h.applicationManagerClient.GetDevice(h.context, &DeviceIdentifier{AppId: appID, DevId: devID})
}

// SetDevice sets a device on the Handler
func (h *ManagerClient) SetDevice(in *Device) error {
	h.RLock()
	defer h.RUnlock()
	_, err := h.applicationManagerClient.SetDevice(h.context, in)
	return err
}

// DeleteDevice deletes a device from the Handler
func (h *ManagerClient) DeleteDevice(appID string, devID string) error {
	h.RLock()
	defer h.RUnlock()
	_, err := h.applicationManagerClient.DeleteDevice(h.context, &DeviceIdentifier{AppId: appID, DevId: devID})
	return err
}

// GetDevicesForApplication retrieves all devices for an application from the Handler
func (h *ManagerClient) GetDevicesForApplication(appID string) (devices []*Device, err error) {
	h.RLock()
	defer h.RUnlock()
	res, err := h.applicationManagerClient.GetDevicesForApplication(h.context, &ApplicationIdentifier{AppId: appID})
	if err != nil {
		return nil, err
	}
	for _, dev := range res.Devices {
		devices = append(devices, dev)
	}
	return
}

// Close closes the client
func (h *ManagerClient) Close() error {
	return h.conn.Close()
}
