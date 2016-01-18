// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/brocaar/lorawan"
	"io/ioutil"
)

// RXPKFromConf read an input json file and parse it into a list of RXPK packets
func RXPKFromConf(filename string) ([]semtech.RXPK, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var conf []struct {
		DevAddr  string   `json:"dev_addr"`
		Metadata metadata `json:"metadata"`
		Payload  string   `json:"payload"`
	}
	err = json.Unmarshal(content, &conf)
	if err != nil {
		return nil, err
	}
	var rxpks []semtech.RXPK
	for _, c := range conf {
		rxpk := semtech.RXPK(c.Metadata)
		rawAddr, err := hex.DecodeString(c.DevAddr)
		if err != nil {
			return nil, err
		}
		var devAddr lorawan.DevAddr
		copy(devAddr[:], rawAddr)
		rxpk.Data = pointer.String(generateData(c.Payload, devAddr))
		rxpks = append(rxpks, rxpk)
	}

	return rxpks, nil
}

type metadata semtech.RXPK // metadata is just an alias used to mislead the UnmarshalJSON

// UnmarshalJSON implements the json.Unmarshal interface
func (m *metadata) UnmarshalJSON(raw []byte) error {
	if m == nil {
		return fmt.Errorf("Cannot unmarshal in nil metadata")
	}
	payload := new(semtech.Payload)
	rawPayload := append(append([]byte(`{"rxpk":[`), raw...), []byte(`]}`)...)
	err := json.Unmarshal(rawPayload, payload)
	if err != nil {
		return err
	}
	if len(payload.RXPK) < 1 {
		return fmt.Errorf("Unable to interpret raw bytes as valid metadata")
	}
	*m = metadata(payload.RXPK[0])
	return nil
}
