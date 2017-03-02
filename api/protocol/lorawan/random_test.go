// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package lorawan

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/api"
	s "github.com/smartystreets/assertions"
)

func TestRandomizers(t *testing.T) {
	for name, msg := range map[string]interface{}{
		"RandomMetadata(FSK)":         RandomMetadata(Modulation_FSK),
		"RandomTxConfiguration(FSK)":  RandomTxConfiguration(Modulation_FSK),
		"RandomMetadata(LORA)":        RandomMetadata(Modulation_LORA),
		"RandomTxConfiguration(LORA)": RandomTxConfiguration(Modulation_LORA),
	} {
		t.Run(name, func(t *testing.T) {
			if v, ok := msg.(api.Validator); ok {
				t.Run("Validate", func(t *testing.T) {
					s.New(t).So(v.Validate(), s.ShouldBeNil)
				})
			}
		})
	}

	for name, payload := range map[string][]byte{
		"RandomPayload(JOIN_REQUEST)":     RandomPayload(MType_JOIN_REQUEST),
		"RandomPayload(JOIN_ACCEPT)":      RandomPayload(MType_JOIN_ACCEPT),
		"RandomPayload(CONFIRMED_UP)":     RandomPayload(MType_CONFIRMED_UP),
		"RandomPayload(CONFIRMED_DOWN)":   RandomPayload(MType_CONFIRMED_DOWN),
		"RandomPayload(UNCONFIRMED_UP)":   RandomPayload(MType_UNCONFIRMED_UP),
		"RandomPayload(UNCONFIRMED_DOWN)": RandomPayload(MType_UNCONFIRMED_DOWN),
		"RandomPayload()":                 RandomPayload(),
		"RandomDownlinkPayload()":         RandomDownlinkPayload(),
		"RandomUplinkPayload()":           RandomUplinkPayload(),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := MessageFromPHYPayloadBytes(payload)
			s.New(t).So(err, s.ShouldBeNil)
		})
	}
}
