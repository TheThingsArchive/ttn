// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"reflect"
	"time"

	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/fatih/structs"
)

const currentDBVersion = "2.4.1"

type DevNonce [2]byte
type AppNonce [3]byte

// Options for the device
type Options struct {
	ActivationConstraints string `json:"activation_constraints,omitempty"` // Activation Constraints (public/local/private)
	DisableFCntCheck      bool   `json:"disable_fcnt_check,omitemtpy"`     // Disable Frame counter check (insecure)
	Uses32BitFCnt         bool   `json:"uses_32_bit_fcnt,omitemtpy"`       // Use 32-bit Frame counters
}

// Device contains the state of a device
type Device struct {
	old *Device

	DevEUI types.DevEUI `redis:"dev_eui"`
	AppEUI types.AppEUI `redis:"app_eui"`
	AppID  string       `redis:"app_id"`
	DevID  string       `redis:"dev_id"`

	Description string `redis:"description"`

	Latitude  float32 `redis:"latitude"`
	Longitude float32 `redis:"longitude"`
	Altitude  int32   `redis:"altitude"`

	Options Options `redis:"options"`

	AppKey        types.AppKey `redis:"app_key"`
	UsedDevNonces []DevNonce   `redis:"used_dev_nonces"`
	UsedAppNonces []AppNonce   `redis:"used_app_nonces"`

	DevAddr types.DevAddr `redis:"dev_addr"`
	NwkSKey types.NwkSKey `redis:"nwk_s_key"`
	AppSKey types.AppSKey `redis:"app_s_key"`
	FCntUp  uint32        `redis:"f_cnt_up"` // Only used to detect retries

	CurrentDownlink *types.DownlinkMessage `redis:"current_downlink"`

	CreatedAt time.Time `redis:"created_at"`
	UpdatedAt time.Time `redis:"updated_at"`
}

// StartUpdate stores the state of the device
func (d *Device) StartUpdate() {
	old := *d
	d.old = &old
}

// DBVersion of the model
func (d *Device) DBVersion() string {
	return currentDBVersion
}

// ChangedFields returns the names of the changed fields since the last call to StartUpdate
func (d Device) ChangedFields() (changed []string) {
	new := structs.New(d)
	fields := new.Names()
	if d.old == nil {
		return fields
	}
	old := structs.New(*d.old)

	for _, field := range new.Fields() {
		if !field.IsExported() || field.Name() == "old" {
			continue
		}
		if !reflect.DeepEqual(field.Value(), old.Field(field.Name()).Value()) {
			changed = append(changed, field.Name())
		}
	}
	return
}

// GetLoRaWAN returns a LoRaWAN Device proto
func (d Device) GetLoRaWAN() *pb_lorawan.Device {
	dev := &pb_lorawan.Device{
		AppId:                 d.AppID,
		DevId:                 d.DevID,
		AppEui:                &d.AppEUI,
		DevEui:                &d.DevEUI,
		DevAddr:               &d.DevAddr,
		NwkSKey:               &d.NwkSKey,
		DisableFCntCheck:      d.Options.DisableFCntCheck,
		Uses32BitFCnt:         d.Options.Uses32BitFCnt,
		ActivationConstraints: d.Options.ActivationConstraints,
	}
	return dev
}
