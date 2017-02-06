// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package migrate

import (
	"github.com/TheThingsNetwork/ttn/core/storage"
)

var deviceMigrations = map[string]func(string) storage.MigrateFunction{}

// DeviceMigrations filled with the prefix
func DeviceMigrations(prefix string) map[string]storage.MigrateFunction {
	funcs := make(map[string]storage.MigrateFunction)
	for v, f := range deviceMigrations {
		funcs[v] = f(prefix)
	}
	return funcs
}
