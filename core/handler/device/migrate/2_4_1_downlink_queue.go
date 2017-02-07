// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package migrate

import (
	"strings"

	"github.com/TheThingsNetwork/ttn/core/storage"
	redis "gopkg.in/redis.v5"
)

// DownlinkQueue migration from 2.4.1 to 2.5.0
func DownlinkQueue(prefix string) storage.MigrateFunction {
	return func(client *redis.Client, key string, obj map[string]string) (string, map[string]string, error) {
		var err error
		nextDownlink, ok := obj["next_downlink"]
		if ok {
			delete(obj, "next_downlink")
			scheduleKey := prefix + ":downlink" + strings.TrimPrefix(key, prefix+":device")
			err = client.LPush(scheduleKey, nextDownlink).Err()
		}
		return "2.4.2", obj, err
	}
}

func init() {
	deviceMigrations["2.4.1"] = DownlinkQueue
}
