// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package migrate

import (
	"encoding/json"
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/storage"
	redis "gopkg.in/redis.v5"
)

type includeOptionsOld struct {
	ActivationConstraints string `json:"activation_constraints,omitempty"`
	DisableFCntCheck      bool   `json:"disable_fcnt_check,omitemtpy"`
	Uses32BitFCnt         bool   `json:"uses_32_bit_fcnt,omitemtpy"`
}

// IncludeOptions migration from 2.4.2 to 2.8.0
func IncludeOptions(prefix string) storage.MigrateFunction {
	return func(client *redis.Client, key string, obj map[string]string) (string, map[string]string, error) {
		if optionsStr, ok := obj["options"]; ok {
			var options includeOptionsOld
			json.Unmarshal([]byte(optionsStr), &options)
			delete(obj, "options")
			obj["options.activation_constraints"] = options.ActivationConstraints
			obj["options.disable_fcnt_check"] = fmt.Sprint(options.DisableFCntCheck)
			obj["options.uses_32_bit_fcnt"] = fmt.Sprint(options.Uses32BitFCnt)
		}
		return "2.8.0", obj, nil
	}
}

func init() {
	deviceMigrations["2.4.2"] = IncludeOptions
}
