// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package migrate

import (
	"github.com/TheThingsNetwork/ttn/core/storage"
	redis "gopkg.in/redis.v5"
)

// MoveFrequencyPlan migration from 2.8.0 to 2.8.1
func MoveFrequencyPlan(prefix string) storage.MigrateFunction {
	return func(client *redis.Client, key string, obj map[string]string) (string, map[string]string, error) {
		if fp, ok := obj["adr.band"]; ok {
			delete(obj, "adr.band")
			obj["options.frequency_plan"] = fp
		}
		return "2.8.1", obj, nil
	}
}

func init() {
	deviceMigrations["2.8.0"] = MoveFrequencyPlan
}
