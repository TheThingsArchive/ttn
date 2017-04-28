// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"math/rand"

	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/utils/random"
)

// RandomLocation returns randomly generated gateway location.
// Used for testing.
func RandomLocation() (gps *GPSMetadata) {
	return &GPSMetadata{
		Longitude: rand.Float32(),
		Latitude:  rand.Float32(),
		Altitude:  rand.Int31(),
	}
}

// RandomRxMetadata returns randomly generated gateway rx metadata.
// Used for testing.
func RandomRxMetadata() *RxMetadata {
	return &RxMetadata{
		GatewayId:      random.ID(),
		Timestamp:      rand.Uint32(),
		Time:           rand.Int63(),
		RfChain:        rand.Uint32(),
		Channel:        rand.Uint32(),
		Frequency:      uint64(random.Freq() * 1000000),
		Rssi:           float32(random.Rssi()),
		Snr:            random.Lsnr(),
		Gps:            RandomLocation(),
		GatewayTrusted: random.Bool(),
	}
}

// RandomTxConfiguration returns randomly generated gateway tx configuration.
// Used for testing.
func RandomTxConfiguration() *TxConfiguration {
	return &TxConfiguration{
		Timestamp:             rand.Uint32(),
		RfChain:               rand.Uint32(),
		Frequency:             uint64(random.Freq() * 1000000),
		Power:                 rand.Int31(),
		FrequencyDeviation:    rand.Uint32(),
		PolarizationInversion: random.Bool(),
	}
}

// RandomStatus returns randomly generated gateway status.
// Used for testing.
func RandomStatus() *Status {
	return &Status{
		Gps:            RandomLocation(),
		Timestamp:      rand.Uint32(),
		Time:           rand.Int63(),
		Ip:             []string{"42.42.42.42"},
		Platform:       random.String(rand.Intn(10)),
		ContactEmail:   fmt.Sprintf("%s@%s.%s", random.String(rand.Intn(10)), random.String(rand.Intn(10)), random.String(rand.Intn(3))),
		Description:    random.String(rand.Intn(10)),
		FrequencyPlan:  lorawan.FrequencyPlan(rand.Intn(4)).String(),
		Bridge:         random.ID(),
		Router:         random.ID(),
		Rtt:            rand.Uint32(),
		RxIn:           rand.Uint32(),
		RxOk:           rand.Uint32(),
		TxIn:           rand.Uint32(),
		TxOk:           rand.Uint32(),
		GatewayTrusted: random.Bool(),
	}
}
