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

func RandomPayload(mType ...MType) []byte {
	m := &Message{}
	m.initMACPayload()

	var mTypeVal MType

	if len(mType) > 0 {
		mTypeVal = mType[0]
	} else {
		switch rand.Intn(6) {
		case 0:
			mTypeVal = MType_JOIN_ACCEPT
		case 1:
			mTypeVal = MType_JOIN_REQUEST
		case 2:
			mTypeVal = MType_UNCONFIRMED_UP
		case 3:
			mTypeVal = MType_UNCONFIRMED_DOWN
		case 4:
			mTypeVal = MType_CONFIRMED_UP
		case 5:
			mTypeVal = MType_CONFIRMED_DOWN
		}
	}

	m.MHDR.MType = mTypeVal

	switch m.MHDR.MType {
	case MType_JOIN_REQUEST:
		payload := JoinRequestPayloadFromPayload(&lorawan.JoinRequestPayload{
			AppEUI:   randomAppEUI(),
			DevEUI:   randomDevEUI(),
			DevNonce: randomDevNonce(),
		})
		m.Payload = &Message_JoinRequestPayload{&payload}
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
		m.Payload = &Message_JoinAcceptPayload{&payload}
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
		m.Payload = &Message_MacPayload{&payload}
	}

	return m.PHYPayloadBytes()
}

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
