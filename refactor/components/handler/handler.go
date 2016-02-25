// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	. "github.com/TheThingsNetwork/ttn/refactor"
)

// component implements the core.Component interface
type component struct{}

// New construct a new Handler component from ...
func New() (Component, error) {
	return nil, nil
}

// Register implements the core.Component interface
func (h component) Register(reg Registration, an AckNacker) error {
	return nil
}

// HandleUp implements the core.Component interface
func (h component) HandleUp(p []byte, an AckNacker, up Adapter) error {
	return nil
}

// HandleDown implements the core.Component interface
func (h component) HandleDown(p []byte, an AckNacker, down Adapter) error {
	return nil
}
