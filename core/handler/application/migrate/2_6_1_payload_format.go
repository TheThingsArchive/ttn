// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package migrate

import (
	"github.com/TheThingsNetwork/ttn/core/storage"
	redis "gopkg.in/redis.v5"
)

// AddPayloadFormat migration from 2.0.0 to 2.6.1
func AddPayloadFormat(prefix string) storage.MigrateFunction {
	return func(client *redis.Client, key string, obj map[string]string) (string, map[string]string, error) {
		any := false
		if decoder, ok := obj["decoder"]; ok {
			delete(obj, "decoder")
			obj["custom_decoder"] = decoder
			any = true
		}
		if converter, ok := obj["converter"]; ok {
			delete(obj, "converter")
			obj["custom_converter"] = converter
			any = true
		}
		if validator, ok := obj["validator"]; ok {
			delete(obj, "validator")
			obj["custom_validator"] = validator
			any = true
		}
		if encoder, ok := obj["encoder"]; ok {
			delete(obj, "encoder")
			obj["custom_encoder"] = encoder
			any = true
		}
		if any {
			obj["payload_format"] = "custom"
		}
		return "2.6.2", obj, nil
	}
}

func init() {
	applicationMigrations["2.4.1"] = AddPayloadFormat
}
