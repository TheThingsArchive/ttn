// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collection

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// DataStorage provides methods to select and save data
type DataStorage interface {
	Save(appEUI types.AppEUI, devEUI types.DevEUI, t time.Time, fields map[string]interface{}) error
	Close() error
}
