// Copyright © 2015 Matthias Benkort <matthias.benkort@gmail.com>
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package protocol

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

// -----------------------------------------------------------------
// ------------------------- Marshal (packet *Packet) ([]byte, error)
// -----------------------------------------------------------------

// ---------- PUSH_DATA

// ---------- PUSH_ACK
func checkMarshalPUSH_ACK(packet *Packet) error {
    raw, err := Marshal(packet)

    if err != nil {
        return err
    }

    if len(raw) != 4 {
        return errors.New(fmt.Printf("Invalid raw sequence length: %d", len(raw)))
    }

    if raw[0] != packet.Version {
        return errors.New(fmt.Printf("Invalid raw version: %x", raw[0]))
    }

    if !bytes.Equal(raw[1:3], packet.Token) {
        return errors.New(fmt.Printf("Invalid raw token: %x", raw[1:3]))
    }

    if raw[3] != packet.Identifier {
        return errors.New(fmt.Printf("Invalid raw identifier: %x", raw[3]))
    }

    return err
}

// Marshal a basic push_ack packet
func TestMarshalPUSH_ACK1(*t testing.T) {
    packet := &Packet{
        Version: VERSION,
        Token: []byte{0xAA, 0x14},
        Identifier: PUSH_ACK,
        GatewayId: nil,
        Payload: nil,
    }
    if err := checkMarshalPUSH_ACK(packet); err != nil {
        t.Errorf("Failed to marshal packet: %v", err)
    }
}

// Marshal a push_ack packet with extra useless gatewayId
func TestMarshalPUSH_ACK2(*t testing.T) {
    packet := &Packet{
        Version: VERSION,
        Token: []byte{0xAA, 0x14},
        Identifier: PUSH_ACK,
        GatewayId: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
        Payload: nil,
    }
    if err := checkMarshalPUSH_ACK(packet); err != nil {
        t.Errorf("Failed to marshal packet: %v", err)
    }
}

// Marshal a push_ack packet with extra useless Payload
func TestMarshalPUSH_ACK3(*t testing.T) {
    payload := &Payload{
        Stat: &Stat{
            Rxfw: 14,
            Rxnb: 14,
            Rxok: 14,
        },
    }
    packet := &Packet{
        Version: VERSION,
        Token: []byte{0xAA, 0x14},
        Identifier: PUSH_ACK,
        GatewayId: nil,
        Payload: payload,
    }
    if err := checkMarshalPUSH_ACK(packet); err != nil {
        t.Errorf("Failed to marshal packet: %v", err)
    }
}

// Marshal a push_ack with extra useless gatewayId and payload
func TestMarshalPUSH_ACK4(*t testing.T) {
    payload := &Payload{
        Stat: &Stat{
            Rxfw: 14,
            Rxnb: 14,
            Rxok: 14,
        },
    }
    packet := &Packet{
        Version: VERSION,
        Token: []byte{0xAA, 0x14},
        Identifier: PUSH_ACK,
        GatewayId: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
        Payload: payload,
    }
    if err := checkMarshalPUSH_ACK(packet); err != nil {
        t.Errorf("Failed to marshal packet: %v", err)
    }
}

// Marshal a push_ack with an invalid token (too short)
func TestMarshalPUSH_ACK5(*t testing.T) {
    packet := &Packet{
        Version: VERSION,
        Token: []byte{0xAA},
        Identifier: PUSH_ACK,
        GatewayId: nil,
        Payload: nil,
    }
    _, err := Marshal(packet)
    if err == nil {
        t.Errorf("Successfully marshalled an invalid packet")
    }
}

// Marshal a push_ack with an invalid token (too long)
func TestMarshalPUSH_ACK6(*t testing.T) {
    packet := &Packet{
        Version: VERSION,
        Token: []byte{0x9A, 0x7A, 0x7E},
        Identifier: PUSH_ACK,
        GatewayId: nil,
        Payload: nil,
    }
    _, err := Marshal(packet)
    if err == nil {
        t.Errorf("Successfully marshalled an invalid packet")
    }
}

// ---------- PULL_DATA

// ---------- PULL_ACK

// ---------- PULL_RESP

