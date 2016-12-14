// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package otaa

import (
	"crypto/aes"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// CalculateSessionKeys calculates the AppSKey and NwkSKey
// All arguments are MSB-first
func CalculateSessionKeys(appKey types.AppKey, appNonce [3]byte, netID [3]byte, devNonce [2]byte) (appSKey types.AppSKey, nwkSKey types.NwkSKey, err error) {

	buf := make([]byte, 16)
	copy(buf[1:4], reverse(appNonce[:]))
	copy(buf[4:7], reverse(netID[:]))
	copy(buf[7:9], reverse(devNonce[:]))

	block, _ := aes.NewCipher(appKey[:])

	buf[0] = 0x1
	block.Encrypt(nwkSKey[:], buf)
	buf[0] = 0x2
	block.Encrypt(appSKey[:], buf)

	return
}

// reverse is used to convert between MSB-first and LSB-first
func reverse(in []byte) (out []byte) {
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return
}
