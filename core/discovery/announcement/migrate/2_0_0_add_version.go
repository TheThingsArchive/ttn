// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package migrate

import (
	"github.com/TheThingsNetwork/ttn/core/storage"
	redis "gopkg.in/redis.v5"
)

// AddVersion migration from nothing to 2.4.1
func AddVersion(prefix string) storage.MigrateFunction {
	return func(client *redis.Client, key string, obj map[string]string) (string, map[string]string, error) {
		return "2.4.1", obj, nil
	}
}

func init() {
	announcementMigrations[""] = AddVersion
}
