// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package lorawan

import (
	"math/rand"

	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/brocaar/lorawan"
)

func randomDevNonce() [2]byte {
	return [2]byte(random.DevNonce())
}
func randomAppNonce() [3]byte {
	return [3]byte(random.AppNonce())
}
func randomNetID() lorawan.NetID {
	return lorawan.NetID(random.NetID())
}
func randomDevAddr() lorawan.DevAddr {
	return lorawan.DevAddr(random.DevAddr())
}
func randomEUI64() lorawan.EUI64 {
	return lorawan.EUI64(random.EUI64())
}
func randomDevEUI() lorawan.EUI64 {
	return randomEUI64()
}
func randomAppEUI() lorawan.EUI64 {
	return randomEUI64()
}

// RandomPayload returns randomly generated payload.
// Used for testing.
func RandomPayload(mType ...MType) []byte {
	msg := &Message{}
	msg.initMACPayload()

	var mTypeVal MType

	if len(mType) > 0 {
		mTypeVal = mType[0]
	} else {
		mTypeVal = MType(rand.Intn(6))
	}

	msg.MHDR.MType = mTypeVal

	switch msg.MHDR.MType {
	case MType_JOIN_REQUEST:
		payload := JoinRequestPayloadFromPayload(&lorawan.JoinRequestPayload{
			AppEUI:   randomAppEUI(),
			DevEUI:   randomDevEUI(),
			DevNonce: randomDevNonce(),
		})
		msg.Payload = &Message_JoinRequestPayload{&payload}
	case MType_JOIN_ACCEPT:
		payload := JoinAcceptPayloadFromPayload(&lorawan.JoinAcceptPayload{
			AppNonce: randomAppNonce(),
			NetID:    randomNetID(),
			DevAddr:  randomDevAddr(),
			RXDelay:  uint8(rand.Intn(15)),
			DLSettings: lorawan.DLSettings{
				RX1DROffset: uint8(rand.Intn(7)),
				RX2DataRate: uint8(rand.Intn(15)),
			},
		})
		msg.Payload = &Message_JoinAcceptPayload{&payload}
	default:
		payload := MACPayloadFromPayload(&lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: randomDevAddr(),
				FCtrl: lorawan.FCtrl{
					ADR:       random.Bool(),
					ADRACKReq: random.Bool(),
					ACK:       random.Bool(),
					FPending:  random.Bool(),
				},
				FCnt: rand.Uint32(),
			},
		})
		msg.Payload = &Message_MacPayload{&payload}
	}

	return msg.PHYPayloadBytes()
}

// RandomUplinkPayload returns randomly generated uplink payload.
// Used for testing.
func RandomUplinkPayload() []byte {
	switch rand.Intn(3) {
	case 0:
		return RandomPayload(MType_JOIN_REQUEST)
	case 1:
		return RandomPayload(MType_UNCONFIRMED_UP)
	default:
		return RandomPayload(MType_CONFIRMED_UP)
	}
}

// RandomDownlinkPayload returns randomly generated downlink payload.
// Used for testing.
func RandomDownlinkPayload() []byte {
	switch rand.Intn(3) {
	case 0:
		return RandomPayload(MType_JOIN_ACCEPT)
	case 1:
		return RandomPayload(MType_UNCONFIRMED_DOWN)
	default:
		return RandomPayload(MType_CONFIRMED_DOWN)
	}
}

// RandomMetadata returns randomly generated Metadata.
// Used for testing.
func RandomMetadata(modulation ...Modulation) *Metadata {
	md := &Metadata{
		FCnt:       rand.Uint32(),
		CodingRate: random.Codr(),
	}

	if len(modulation) == 1 {
		md.Modulation = modulation[0]
	} else {
		if rand.Int()%2 == 0 {
			md.Modulation = Modulation_LORA
		} else {
			md.Modulation = Modulation_FSK
		}
	}

	switch md.Modulation {
	case Modulation_LORA:
		md.DataRate = random.Datr()
	case Modulation_FSK:
		md.BitRate = rand.Uint32()
	}
	return md
}

// RandomTxConfiguration returns randomly generated TxConfiguration.
// Used for testing.
func RandomTxConfiguration(modulation ...Modulation) *TxConfiguration {
	conf := &TxConfiguration{
		FCnt:       rand.Uint32(),
		CodingRate: random.Codr(),
	}

	if len(modulation) == 1 {
		conf.Modulation = modulation[0]
	} else {
		if rand.Int()%2 == 0 {
			conf.Modulation = Modulation_LORA
		} else {
			conf.Modulation = Modulation_FSK
		}
	}

	switch conf.Modulation {
	case Modulation_LORA:
		conf.DataRate = random.Datr()
	case Modulation_FSK:
		conf.BitRate = rand.Uint32()
	}
	return conf
}
