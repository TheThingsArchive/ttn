// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import "fmt"

var ErrInvalidRegistration = fmt.Errorf("Invalid registration")
var ErrDeviceNotFound = fmt.Errorf("Device not found")
var ErrInvalidPacket = fmt.Errorf("The given packet is invalid")
var ErrBadOptions = fmt.Errorf("Invalid supplied options")
var ErrNotInitialized = fmt.Errorf("Illegal operation call on non initialized component")
var ErrEntryExpired = fmt.Errorf("An entry exists but has expired")
var ErrAlreadyExists = fmt.Errorf("An entry already exists for that device")
