// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"reflect"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/fatih/structs"
)

type DevNonce [2]byte
type AppNonce [3]byte

// Device contains the state of a device
type Device struct {
	old           *Device
	DevEUI        types.DevEUI          `redis:"dev_eui"`
	AppEUI        types.AppEUI          `redis:"app_eui"`
	AppID         string                `redis:"app_id"`
	DevID         string                `redis:"dev_id"`
	DevAddr       types.DevAddr         `redis:"dev_addr"`
	AppKey        types.AppKey          `redis:"app_key"`
	UsedDevNonces []DevNonce            `redis:"used_dev_nonces"`
	UsedAppNonces []AppNonce            `redis:"used_app_nonces"`
	NwkSKey       types.NwkSKey         `redis:"nwk_s_key"`
	AppSKey       types.AppSKey         `redis:"app_s_key"`
	NextDownlink  *mqtt.DownlinkMessage `redis:"next_downlink"`

	CreatedAt time.Time `redis:"created_at"`
	UpdatedAt time.Time `redis:"updated_at"`
}

// StartUpdate stores the state of the device
func (d *Device) StartUpdate() {
	old := *d
	d.old = &old
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
