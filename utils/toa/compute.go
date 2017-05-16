// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package toa

import (
	"errors"
	"time"

	"github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
)

type hasProtocolMetadata interface {
	GetProtocolMetadata() *protocol.RxMetadata
}

type hasProtocolConfiguration interface {
	GetProtocolConfiguration() *protocol.TxConfiguration
}

type hasProtocolConfig interface {
	GetProtocolConfig() *protocol.TxConfiguration
}

type hasPayload interface {
	GetPayload() []byte
}

// DefaultPayloadSize is used in Compute() if the message has an empty payload
var DefaultPayloadSize uint = 64

// Compute the time-on-air for the given message
func Compute(msg interface{}) (time.Duration, error) {
	var payloadSize uint
	var modulation lorawan.Modulation
	var dataRate, codingRate string
	var bitRate uint32

	switch msg := msg.(type) {
	case hasProtocolConfig:
		lorawan := msg.GetProtocolConfig().GetLorawan()
		if lorawan == nil {
			return 0, errors.New("toa: message does not have lorawan config")
		}
		modulation = lorawan.GetModulation()
		dataRate = lorawan.GetDataRate()
		bitRate = lorawan.GetBitRate()
		codingRate = lorawan.GetCodingRate()
	case hasProtocolConfiguration:
		lorawan := msg.GetProtocolConfiguration().GetLorawan()
		if lorawan == nil {
			return 0, errors.New("toa: message does not have lorawan configuration")
		}
		modulation = lorawan.GetModulation()
		dataRate = lorawan.GetDataRate()
		bitRate = lorawan.GetBitRate()
		codingRate = lorawan.GetCodingRate()
	case hasProtocolMetadata:
		lorawan := msg.GetProtocolMetadata().GetLorawan()
		if lorawan == nil {
			return 0, errors.New("toa: message does not have lorawan metadata")
		}
		modulation = lorawan.GetModulation()
		dataRate = lorawan.GetDataRate()
		bitRate = lorawan.GetBitRate()
		codingRate = lorawan.GetCodingRate()
	}

	if msg, ok := msg.(hasPayload); ok {
		payloadSize = uint(len(msg.GetPayload()))
	}
	if payloadSize == 0 {
		payloadSize = DefaultPayloadSize
	}

	switch modulation {
	case lorawan.Modulation_LORA:
		return ComputeLoRa(payloadSize, dataRate, codingRate)
	case lorawan.Modulation_FSK:
		return ComputeFSK(payloadSize, int(bitRate))
	}

	return 0, errors.New("toa: could not compute time-on-air")
}
