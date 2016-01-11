package core

import (
	"fmt"
	"github.com/thethingsnetwork/core/lorawan"
)

func (p Packet) DevAddr() (lorawan.DevAddr, error) {
	if p.Payload.MACPayload == nil {
		return lorawan.DevAddr{}, fmt.Errorf("lorawan: MACPayload should not be empty")
	}

	macpayload, ok := p.Payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return lorawan.DevAddr{}, fmt.Errorf("lorawan: unable to get address of a join message")
	}

	return macpayload.FHDR.DevAddr, nil
}

func (p Packet) String() string {
	str := "Packet {"
	if p.Metadata != nil {
		str += fmt.Sprintf("\n\t%s}", p.Metadata.String())
	}
	str += fmt.Sprintf("\n\tPayload%+v\n}", p.Payload)
	return str
}
