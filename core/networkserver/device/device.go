// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"reflect"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/fatih/structs"
)

const currentDBVersion = "2.4.1"

// Options for the specified device
type Options struct {
	ActivationConstraints string `json:"activation_constraints,omitempty"` // Activation Constraints (public/local/private)
	DisableFCntCheck      bool   `json:"disable_fcnt_check,omitemtpy"`     // Disable Frame counter check (insecure)
	Uses32BitFCnt         bool   `json:"uses_32_bit_fcnt,omitemtpy"`       // Use 32-bit Frame counters
}

// Device contains the state of a device
type Device struct {
	old *Device

	DevEUI   types.DevEUI  `redis:"dev_eui"`
	AppEUI   types.AppEUI  `redis:"app_eui"`
	AppID    string        `redis:"app_id"`
	DevID    string        `redis:"dev_id"`
	DevAddr  types.DevAddr `redis:"dev_addr"`
	NwkSKey  types.NwkSKey `redis:"nwk_s_key"`
	FCntUp   uint32        `redis:"f_cnt_up"`
	FCntDown uint32        `redis:"f_cnt_down"`
	LastSeen time.Time     `redis:"last_seen"`
	Options  Options       `redis:"options"`
	ADR      ADRSettings   `redis:"adr,include"`

	CreatedAt time.Time `redis:"created_at"`
	UpdatedAt time.Time `redis:"updated_at"`
}

// ADRSettings contains the (desired) settings for a device that uses ADR
type ADRSettings struct {
	Band   string `redis:"band"`
	Margin int    `redis:"margin"`

	// Indicates whether the NetworkServer should send a LinkADRReq when possible
	SendReq bool `redis:"send_req,omitempty"`
	Failed  int  `redis:"failed,omitempty"` // number of failed ADR attempts

	// Desired Settings:
	DataRate string `redis:"data_rate,omitempty"`
	TxPower  int    `redis:"tx_power,omitempty"`
	NbTrans  int    `redis:"nb_trans,omitempty"`
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

	if len(changed) == 1 && changed[0] == "UpdatedAt" {
		return []string{}
	}

	return
}
