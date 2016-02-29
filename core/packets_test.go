// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"math/rand"
	"testing"

	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/assertions"
)

func randBytes(n int) []byte {
	bytes := make([]byte, n)
	for i := range bytes {
		bytes[i] = byte(rand.Intn(255))
	}
	return bytes
}

func newEUI() lorawan.EUI64 {
	devEUI := [8]byte{}
	copy(devEUI[:], randBytes(8))
	return devEUI
}

func simplePayload() (payload lorawan.PHYPayload, devAddr lorawan.DevAddr, key lorawan.AES128Key) {
	copy(devAddr[:], randBytes(4))
	copy(key[:], randBytes(16))

	payload = newPayload(devAddr, []byte("PLD123"), key, key)
	return
}

func newPayload(devAddr lorawan.DevAddr, data []byte, appSKey lorawan.AES128Key, nwkSKey lorawan.AES128Key) lorawan.PHYPayload {
	uplink := true

	macPayload := lorawan.NewMACPayload(uplink)
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: devAddr,
		FCtrl: lorawan.FCtrl{
			ADR:       false,
			ADRACKReq: false,
			ACK:       false,
		},
		FCnt: 1,
	}
	macPayload.FPort = 10
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: data}}

	if err := macPayload.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	payload := lorawan.NewPHYPayload(uplink)
	payload.MHDR = lorawan.MHDR{
		MType: lorawan.ConfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	payload.MACPayload = macPayload

	if err := payload.SetMIC(nwkSKey); err != nil {
		panic(err)
	}

	return payload
}

func marshalUnmarshal(t *testing.T, input Packet) interface{} {
	binary, err := input.MarshalBinary()
	if err != nil {
		Ko(t, "Unexpected error {%s}.", err)
	}

	gOutput, err := UnmarshalPacket(binary)
	if err != nil {
		Ko(t, "Unexpected error {%s}.", err)
	}

	return gOutput
}

func TestPacket(t *testing.T) {
	input := basegpacket{
		metadata: []Metadata{
			Metadata{
				Codr: pointer.String("4/6"),
			},
		},
	}

	binary, _ := input.Marshal()

	output := basegpacket{
		metadata: []Metadata{},
	}

	_, err := output.Unmarshal(binary)
	if err != nil {
		Ko(t, "Unexpected error {%s}.", err)
	}
}

func TestRPacket(t *testing.T) {
	a := New(t)

	payload, devAddr, _ := simplePayload()
	gwEUI := []byte{}
	copy(gwEUI[:], randBytes(8))

	input, _ := NewRPacket(payload, gwEUI, Metadata{})

	gOutput := marshalUnmarshal(t, input)

	output, ok := gOutput.(RPacket)
	if !ok {
		Ko(t, "Didn't get an RPacket back")
	}

	a.So(output.Payload(), ShouldResemble, payload)
	a.So(output.GatewayId(), ShouldResemble, gwEUI)
	a.So(output.Metadata(), ShouldResemble, Metadata{})
	outputDevEUI := output.DevEUI()
	a.So(outputDevEUI[4:], ShouldResemble, devAddr[:])
}

func TestSPacket(t *testing.T) {
	// Nope
}

func TestBPacket(t *testing.T) {
	a := New(t)

	payload, _, key := simplePayload()
	input, _ := NewBPacket(payload, Metadata{})

	gOutput := marshalUnmarshal(t, input)

	output, ok := gOutput.(BPacket)
	if !ok {
		Ko(t, "Didn't get an BPacket back")
	}

	a.So(output.Payload(), ShouldResemble, payload)
	a.So(output.Metadata(), ShouldResemble, Metadata{})
	outputValidateMIC, _ := output.ValidateMIC(key)
	a.So(outputValidateMIC, ShouldBeTrue)
	a.So(output.Commands(), ShouldBeEmpty)
}

func TestHPacket(t *testing.T) {
	a := New(t)

	appEUI := newEUI()
	devEUI := newEUI()
	payload, _, key := simplePayload()

	input, _ := NewHPacket(appEUI, devEUI, payload, Metadata{})

	gOutput := marshalUnmarshal(t, input)

	output, ok := gOutput.(HPacket)
	if !ok {
		Ko(t, "Didn't get an HPacket back")
	}

	a.So(output.AppEUI().String(), ShouldEqual, appEUI.String())
	a.So(output.DevEUI().String(), ShouldEqual, devEUI.String())
	outPayload, _ := output.Payload(key)
	a.So(string(outPayload), ShouldResemble, "PLD123")
	a.So(output.Metadata(), ShouldResemble, Metadata{})
	a.So(output.FCnt(), ShouldEqual, 1)
}

func TestAPacket(t *testing.T) {
	a := New(t)

	appEUI := newEUI()
	devEUI := newEUI()
	payload := []byte("PLD123")

	input, _ := NewAPacket(appEUI, devEUI, payload, []Metadata{})

	gOutput := marshalUnmarshal(t, input)

	output, ok := gOutput.(APacket)
	if !ok {
		Ko(t, "Didn't get an APacket back")
	}

	a.So(output.Payload(), ShouldResemble, payload)
	a.So(output.DevEUI().String(), ShouldEqual, devEUI.String())
	a.So(output.Payload(), ShouldResemble, payload)
	a.So(output.Metadata(), ShouldBeEmpty)
}

func TestJPacket(t *testing.T) {
	a := New(t)

	appEUI := newEUI()
	devEUI := newEUI()
	devNonce := [2]byte{}
	copy(devEUI[:], randBytes(2))

	input := NewJPacket(appEUI, devEUI, devNonce, Metadata{})

	gOutput := marshalUnmarshal(t, input)

	output, ok := gOutput.(JPacket)
	if !ok {
		Ko(t, "Didn't get an JPacket back")
	}

	a.So(output.AppEUI().String(), ShouldEqual, appEUI.String())
	a.So(output.DevEUI().String(), ShouldEqual, devEUI.String())
	a.So(output.DevNonce(), ShouldEqual, devNonce)
	a.So(output.Metadata(), ShouldResemble, Metadata{})
}

func TestCPacket(t *testing.T) {
	a := New(t)

	appEUI := newEUI()
	devEUI := newEUI()
	payload := []byte("PLD123")

	nwkSKey := [16]byte{}
	copy(devEUI[:], randBytes(16))

	input, _ := NewCPacket(appEUI, devEUI, payload, nwkSKey)

	gOutput := marshalUnmarshal(t, input)

	output, ok := gOutput.(CPacket)
	if !ok {
		Ko(t, "Didn't get an CPacket back")
	}

	a.So(output.AppEUI().String(), ShouldEqual, appEUI.String())
	a.So(output.DevEUI().String(), ShouldEqual, devEUI.String())
	a.So(output.Payload(), ShouldResemble, payload)
	outputNwkSKey := output.NwkSKey()
	a.So(outputNwkSKey[:], ShouldResemble, nwkSKey[:])
}
