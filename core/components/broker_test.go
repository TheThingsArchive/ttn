// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestBrokerHandleup(t *testing.T) {
	devices := []device{
		{
			DevAddr: [4]byte{1, 2, 3, 4},
			AppSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			NwkSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			DevAddr: [4]byte{0, 0, 0, 2},
			AppSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 8, 8, 8, 8, 8, 8},
			NwkSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 8, 8, 8, 8, 8, 8},
		},
		{
			DevAddr: [4]byte{14, 14, 14, 14},
			AppSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 8, 11, 8, 11, 8, 8},
			NwkSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 8, 10, 11, 8, 8, 8},
		},
		{
			DevAddr: [4]byte{1, 2, 3, 4},
			AppSKey: [16]byte{1, 2, 3, 4, 4, 5, 9, 7, 7, 9, 8, 8, 8, 3, 13, 8},
			NwkSKey: [16]byte{1, 2, 3, 4, 4, 5, 4, 7, 9, 9, 8, 8, 8, 9, 14, 8},
		},
	}

	tests := []struct {
		Desc            string
		KnownRecipients []core.Registration
		Packet          packetShape
		WantRecipients  []core.Recipient
		WantAck         bool
		WantError       error
	}{
		{
			Desc: "0 known | Send #0",
			Packet: packetShape{
				Device: devices[0],
				Data:   "MyData",
			},
			WantRecipients: nil,
			WantAck:        false,
			WantError:      nil,
		},
	}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		broker := genNewBroker(t, test.KnownRecipients)
		packet := genPacketFromShape(test.Packet)

		// Operate
		recipients, ack, err := handleBrokerUp(broker, packet)

		// Check
		checkErrors(t, test.WantError, err)
		checkBrokerAcks(t, test.WantAck, ack)
		checkRecipients(t, test.WantRecipients, recipients)

		if err := broker.db.Close(); err != nil {
			panic(err)
		}
	}
}

// ----- BUILD utilities
func genNewBroker(t *testing.T, knownRecipients []core.Registration) *Broker {
	ctx := GetLogger(t, "Broker")

	db, err := NewBrokerStorage()
	if err != nil {
		panic(err)
	}

	if err := db.Reset(); err != nil {
		panic(err)
	}

	broker := NewBroker(db, ctx)
	if err != nil {
		panic(err)
	}

	for _, registration := range knownRecipients {
		err := broker.Register(registration, voidAckNacker{})
		if err != nil {
			panic(err)
		}
	}

	return broker
}

// ----- OPERATE utilities
func handleBrokerUp(broker core.Broker, packet core.Packet) ([]core.Recipient, *bool, error) {
	adapter := &routerAdapter{}
	an := &brokerAckNacker{}
	err := broker.HandleUp(packet, an, adapter)
	return adapter.Recipients, an.HasAck, err
}

type brokerAckNacker struct {
	HasAck *bool
}

func (an *brokerAckNacker) Ack(packets ...core.Packet) error {
	an.HasAck = new(bool)
	*an.HasAck = true
	return nil
}

func (an *brokerAckNacker) Nack() error {
	an.HasAck = new(bool)
	*an.HasAck = false
	return nil
}

// ----- CHECK utilities
func checkBrokerAcks(t *testing.T, want bool, got *bool) {
	if got == nil {
		Ko(t, "No Ack or Nack was sent")
		return
	}

	expected, notExpected := "ack", "nack"
	if !want {
		expected, notExpected = notExpected, expected
	}
	if want != *got {
		Ko(t, "Expected %s but got %s", expected, notExpected)
		return
	}
	Ok(t, "Check acks")
}
