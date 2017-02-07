// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package application

import (
	"reflect"
	"time"

	"github.com/fatih/structs"
)

const currentDBVersion = "2.4.1"

// Application contains the state of an application
type Application struct {
	old *Application

	AppID string `redis:"app_id"`
	// Decoder is a JavaScript function that accepts the payload as byte array and
	// returns an object containing the decoded values
	Decoder string `redis:"decoder"`
	// Converter is a JavaScript function that accepts the data as decoded by
	// Decoder and returns an object containing the converted values
	Converter string `redis:"converter"`
	// Validator is a JavaScript function that validates the data is converted by
	// Converter and returns a boolean value indicating the validity of the data
	Validator string `redis:"validator"`
	// Encoder is a JavaScript function that encode the data send on Downlink messages
	// Returns an object containing the converted values in []byte
	Encoder string `redis:"encoder"`

	CreatedAt time.Time `redis:"created_at"`
	UpdatedAt time.Time `redis:"updated_at"`
}

// StartUpdate stores the state of the device
func (a *Application) StartUpdate() {
	old := *a
	a.old = &old
}

// DBVersion of the model
func (a *Application) DBVersion() string {
	return currentDBVersion
}

// ChangedFields returns the names of the changed fields since the last call to StartUpdate
func (a Application) ChangedFields() (changed []string) {
	new := structs.New(a)
	fields := new.Names()
	if a.old == nil {
		return fields
	}
	old := structs.New(*a.old)

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
