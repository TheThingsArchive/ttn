// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"fmt"
	"net"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/semtech"
)

// semtechAckNacker represents an AckNacker for a semtech request
type semtechAckNacker struct {
	conn      chan udpMsg    // The adapter downlink connection channel
	recipient core.Recipient // The recipient to reach
}

// Ack implements the core.Adapter interface
func (an semtechAckNacker) Ack(p ...core.Packet) error {
	if len(p) == 0 {
		return nil
	}

	// For the downlink, we have to send a PULL_RESP packet which hold a TXPK
	txpk, err := core.ConvertToTXPK(p[0])
	if err != nil {
		return err
	}

	raw, err := semtech.Packet{
		Version:    semtech.VERSION,
		Identifier: semtech.PULL_RESP,
		Payload:    &semtech.Payload{TXPK: &txpk},
	}.MarshalBinary()

	addr, ok := an.recipient.Address.(*net.UDPAddr)
	if !ok {
		return fmt.Errorf("Recipient address was invalid. Expected UDPAddr but got: %v", an.recipient.Address)
	}
	an.conn <- udpMsg{raw: raw, addr: addr}
	return nil
}

// Ack implements the core.Adapter interface
func (an semtechAckNacker) Nack() error {
	// There's no notion of nack in the semtech protocol. You either reply something or you don't.
	return nil
}
