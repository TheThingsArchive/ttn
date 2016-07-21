package handler

import (
	"os"
	"os/user"
	"sync"

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
	return h.applicationManagerClient.GetApplication(h.getContext(), &ApplicationIdentifier{AppId: appID})
}

// SetApplication sets an application on the Handler
func (h *ManagerClient) SetApplication(in *Application) error {
	_, err := h.applicationManagerClient.SetApplication(h.getContext(), in)
	return err
}

// RegisterApplication registers an application on the Handler
func (h *ManagerClient) RegisterApplication(appID string) error {
	_, err := h.applicationManagerClient.RegisterApplication(h.getContext(), &ApplicationIdentifier{AppId: appID})
	return err
}

// DeleteApplication deletes an application and all its devices from the Handler
func (h *ManagerClient) DeleteApplication(appID string) error {
	_, err := h.applicationManagerClient.DeleteApplication(h.getContext(), &ApplicationIdentifier{AppId: appID})
	return err
}

// GetDevice retrieves a device from the Handler
func (h *ManagerClient) GetDevice(appID string, devID string) (*Device, error) {
	return h.applicationManagerClient.GetDevice(h.getContext(), &DeviceIdentifier{AppId: appID, DevId: devID})
}

// SetDevice sets a device on the Handler
func (h *ManagerClient) SetDevice(in *Device) error {
	_, err := h.applicationManagerClient.SetDevice(h.getContext(), in)
	return err
}

// DeleteDevice deletes a device from the Handler
func (h *ManagerClient) DeleteDevice(appID string, devID string) error {
	_, err := h.applicationManagerClient.DeleteDevice(h.getContext(), &DeviceIdentifier{AppId: appID, DevId: devID})
	return err
}

// GetDevicesForApplication retrieves all devices for an application from the Handler
func (h *ManagerClient) GetDevicesForApplication(appID string) (devices []*Device, err error) {
	res, err := h.applicationManagerClient.GetDevicesForApplication(h.getContext(), &ApplicationIdentifier{AppId: appID})
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
