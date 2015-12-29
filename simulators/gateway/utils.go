// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"bytes"
	"encoding/binary"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"math/rand"
)

func genToken() []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, rand.Uint32())
	return b[0:2]
}

func ackToken(index int, packet semtech.Packet) [4]byte {
	buf := new(bytes.Buffer)
	var id byte
	if err := binary.Write(buf, binary.LittleEndian, uint16(index)); err != nil {
		id = 0xff
	} else {
		id = buf.Bytes()[0]
	}

	var kind byte
	switch packet.Identifier {
	case semtech.PUSH_ACK, semtech.PUSH_DATA:
		kind = 0x1
	case semtech.PULL_ACK, semtech.PULL_DATA, semtech.PULL_RESP:
		kind = 0x2
	}

	return [4]byte{id, kind, packet.Token[0], packet.Token[1]}
}

func generateRssi() int {
	x := float32(rand.Int31()) / float32(2e8)
	return -int(x * x)
}

func generateFreq() float64 {
	// EU 863-870MHz
	return rand.Float64()*7 + 863.0
}
