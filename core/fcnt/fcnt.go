// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package fcnt

const maxUint16 = (1 << 16)

// GetFull calculates the full 32-bit frame counter
func GetFull(full uint32, lsb uint16) uint32 {
	if int(lsb)-int(full) > 0 {
		return uint32(lsb)
	}
	if uint16(full%maxUint16) <= lsb {
		return uint32(lsb) + (full/maxUint16)*maxUint16
	}
	return uint32(lsb) + ((full/maxUint16)+1)*maxUint16
}
