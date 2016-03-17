// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/semtech"
	"testing"
	//	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestInjectMetadata(t *testing.T) {
	{
		Desc(t, "Inject full set of Metadata")

		// Build
		rxpk := new(semtech.RXPK)
		meta := core.Metadata{
			Altitude:    -12,
			CodingRate:  "4/5",
			DataRate:    "SF7BW125",
			DutyRX1:     12,
			DutyRX2:     0,
			Frequency:   868.34,
			Latitude:    12.67,
			Longitude:   33.45,
			Lsnr:        5.2,
			PayloadSize: 14,
			Rssi:        -12,
			Timestamp:   123455,
		}

		// Expectation
		var want = &semtech.RXPK{
			Codr: pointer.String(meta.CodingRate),
			Datr: pointer.String(meta.DataRate),
			Freq: pointer.Float32(meta.Frequency),
			Lsnr: pointer.Float32(meta.Lsnr),
			Size: pointer.Uint32(meta.PayloadSize),
			Rssi: pointer.Int32(meta.Rssi),
			Tmst: pointer.Uint32(meta.Timestamp),
		}

		// Operate
		_ = injectMetadata(rxpk, meta)

		// Check
		Check(t, want, rxpk, "RXPKs")
	}

	// --------------------

	{
		Desc(t, "Partial Set of Metadata")

		// Build
		rxpk := new(semtech.RXPK)
		meta := core.Metadata{
			Altitude:   -12,
			CodingRate: "4/5",
			Rssi:       -12,
			Timestamp:  123455,
		}

		// Expectation
		var want = &semtech.RXPK{
			Codr: pointer.String(meta.CodingRate),
			Rssi: pointer.Int32(meta.Rssi),
			Tmst: pointer.Uint32(meta.Timestamp),
		}

		// Operate
		_ = injectMetadata(rxpk, meta)

		// Check
		Check(t, want, rxpk, "RXPKs")
	}
}

func TestExtractMetadta(t *testing.T) {
	{
		Desc(t, "Extract full set of Metadata")

		// Build
		txpk := semtech.TXPK{
			Codr: pointer.String("4/5"),
			Data: pointer.String("4xvansicvni7bvcxxcvxds=="),
			Datr: pointer.String("SF7BW125"),
			Fdev: pointer.Uint32(42),
			Freq: pointer.Float32(868.45),
			Imme: pointer.Bool(true),
			Ipol: pointer.Bool(false),
			Modu: pointer.String("Lora"),
			Ncrc: pointer.Bool(false),
			Powe: pointer.Uint32(12000),
			Prea: pointer.Uint32(1000),
			Rfch: pointer.Uint32(3),
			Size: pointer.Uint32(14),
			Time: pointer.Time(time.Now()),
			Tmst: pointer.Uint32(23456789),
		}
		meta := new(core.AppMetadata)

		// Expectation
		var want = &core.AppMetadata{
			CodingRate: *txpk.Codr,
			DataRate:   *txpk.Datr,
			Frequency:  *txpk.Freq,
			Timestamp:  *txpk.Tmst,
		}

		// Operate
		_ = extractMetadata(txpk, meta)

		// Check
		Check(t, want, meta, "Metadata")
	}

	// --------------------

	{
		Desc(t, "Extract partial set of Metadata")

		// Build
		txpk := semtech.TXPK{
			Codr: pointer.String("4/5"),
			Datr: pointer.String("SF7BW125"),
			Fdev: pointer.Uint32(42),
			Imme: pointer.Bool(true),
			Ipol: pointer.Bool(false),
			Ncrc: pointer.Bool(false),
			Powe: pointer.Uint32(12000),
			Prea: pointer.Uint32(1000),
			Size: pointer.Uint32(14),
		}
		meta := new(core.AppMetadata)

		// Expectation
		var want = &core.AppMetadata{
			CodingRate: *txpk.Codr,
			DataRate:   *txpk.Datr,
		}

		// Operate
		_ = extractMetadata(txpk, meta)

		// Check
		Check(t, want, meta, "Metadata")
	}

}

func TestNewTXPK(t *testing.T) {

}

func TestNewDataRouterReq(t *testing.T) {

}
