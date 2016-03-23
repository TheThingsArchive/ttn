// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
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
			Modulation: *txpk.Modu,
			RFChain:    *txpk.Rfch,
			Time:       txpk.Time.Format(time.RFC3339Nano),
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

// func newTXPK(resp core.DataRouterRes, ctx log.Interface) (semtech.TXPK, error) {
func TestNewTXPKAndtoLoRaWANPayloadReq(t *testing.T) {
	{
		Desc(t, "Test Valid marshal / unmarshal")

		// Build
		res := core.DataRouterRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl: &core.LoRaWANFCtrl{
							ADR:       false,
							ADRAckReq: false,
							Ack:       false,
							FPending:  false,
							FOptsLen:  nil,
						},
						FOpts: nil,
					},
					FPort:      1,
					FRMPayload: []byte{5, 6, 7, 8},
				},
				MIC: []byte{14, 42, 14, 42},
			},
			Metadata: &core.Metadata{
				CodingRate: "4/5",
				DataRate:   "SF8BW125",
			},
		}
		payload, err := core.NewLoRaWANData(res.Payload, false)
		FatalUnless(t, err)
		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}

		// Expectations
		var wantReq = &core.DataRouterReq{
			Payload:   res.Payload,
			Metadata:  res.Metadata,
			GatewayID: gid,
		}
		var wantErrTXPK *string
		var wantErrReq *string

		// Operate
		txpk, errTXPK := newTXPK(payload, res.Metadata, GetLogger(t, "Convert TXPK"))
		req, errReq := toLoRaWANPayload(semtech.RXPK{
			Codr: txpk.Codr,
			Datr: txpk.Datr,
			Data: txpk.Data,
		}, gid, GetLogger(t, "Convert DataRouterReq"))

		// Check
		CheckErrors(t, wantErrTXPK, errTXPK)
		CheckErrors(t, wantErrReq, errReq)
		Check(t, wantReq, req, "Data Router Requests")
	}

	// --------------------

	{
		Desc(t, "New TXPK with no Metadata")

		// Build
		res := core.DataRouterRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl: &core.LoRaWANFCtrl{
							ADR:       false,
							ADRAckReq: false,
							Ack:       false,
							FPending:  false,
							FOptsLen:  nil,
						},
						FOpts: nil,
					},
					FPort:      1,
					FRMPayload: []byte{5, 6, 7, 8},
				},
				MIC: []byte{14, 42, 14, 42},
			},
			Metadata: nil,
		}
		payload, err := core.NewLoRaWANData(res.Payload, false)
		FatalUnless(t, err)

		// Expectations
		var wantTXPK semtech.TXPK
		var wantErr = ErrStructural

		// Operate
		txpk, errTXPK := newTXPK(payload, res.Metadata, GetLogger(t, "Convert TXPK"))

		// Check
		CheckErrors(t, wantErr, errTXPK)
		Check(t, wantTXPK, txpk, "TXPKs")
	}

	// --------------------

	{
		Desc(t, "Test toLoRaWANPayload with invalid macpayload")

		// Build
		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		payload := lorawan.NewPHYPayload(true)
		payload.MHDR.MType = lorawan.UnconfirmedDataUp
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		payload.MACPayload = pointer.Time(time.Now()) // Something marshable
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)
		rxpk := semtech.RXPK{
			Codr: pointer.String("4/5"),
			Freq: pointer.Float32(867.345),
			Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
		}

		// Expectations
		var wantReq interface{}
		var wantErr = ErrStructural

		// Operate
		req, err := toLoRaWANPayload(rxpk, gid, GetLogger(t, "Convert DataRouterReq"))

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantReq, req, "Data Router Requests")
	}

	// --------------------

	{
		Desc(t, "Test toLoRaWANPayload with no data in rxpk")

		// Build
		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		rxpk := semtech.RXPK{
			Codr: pointer.String("4/5"),
			Freq: pointer.Float32(867.345),
		}

		// Expectations
		var wantReq interface{}
		var wantErr = ErrStructural

		// Operate
		req, err := toLoRaWANPayload(rxpk, gid, GetLogger(t, "Convert DataRouterReq"))

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantReq, req, "Data Router Requests")
	}

	// --------------------

	{
		Desc(t, "Test toLoRaWANPayload with random encoded data in rxpk")

		// Build
		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		rxpk := semtech.RXPK{
			Codr: pointer.String("4/5"),
			Freq: pointer.Float32(867.345),
			Data: pointer.String("$#%$^ffg"),
		}

		// Expectations
		var wantReq interface{}
		var wantErr = ErrStructural

		// Operate
		req, err := toLoRaWANPayload(rxpk, gid, GetLogger(t, "Convert DataRouterReq"))

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantReq, req, "Data Router Requests")
	}

	// --------------------

	{
		Desc(t, "Test toLoRaWANPayload with not base64 data")

		// Build
		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		rxpk := semtech.RXPK{
			Codr: pointer.String("4/5"),
			Freq: pointer.Float32(867.345),
			Data: pointer.String("Patate"),
		}

		// Expectations
		var wantReq interface{}
		var wantErr = ErrStructural

		// Operate
		req, err := toLoRaWANPayload(rxpk, gid, GetLogger(t, "Convert DataRouterReq"))

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantReq, req, "Data Router Requests")
	}

	// --------------------

	{
		Desc(t, "Test toLoRaWANPayload with mac commands in fopts")

		// Build
		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		payload := lorawan.NewPHYPayload(false)
		payload.MHDR.MType = lorawan.ConfirmedDataUp
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		macpayload := lorawan.NewMACPayload(false)
		macpayload.FPort = new(uint8)
		*macpayload.FPort = 1

		macpayload.FHDR.FOpts = []lorawan.MACCommand{
			lorawan.MACCommand{
				CID: lorawan.DutyCycleReq,
				Payload: &lorawan.DutyCycleReqPayload{
					MaxDCCycle: 14,
				},
			},
		}
		macpayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{1, 2, 3}}}
		payload.MACPayload = macpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)
		rxpk := semtech.RXPK{
			Codr: pointer.String("4/5"),
			Freq: pointer.Float32(867.345),
			Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
		}

		// Expectations
		var wantErr *string

		// Operate
		_, err = toLoRaWANPayload(rxpk, gid, GetLogger(t, "Convert DataRouterReq"))

		// Check
		CheckErrors(t, wantErr, err)
	}

	// --------------------

	{
		Desc(t, "Test toLoRaWANPayload with unhandled mtype")

		// Build
		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		payload := lorawan.NewPHYPayload(false)
		payload.MHDR.MType = lorawan.UnconfirmedDataDown
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		macpayload := lorawan.NewMACPayload(false)
		macpayload.FPort = new(uint8)
		*macpayload.FPort = 1

		macpayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{1, 2, 3}}}
		payload.MACPayload = macpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)
		rxpk := semtech.RXPK{
			Codr: pointer.String("4/5"),
			Freq: pointer.Float32(867.345),
			Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
		}

		// Expectations
		var wantErr = ErrStructural

		// Operate
		_, err = toLoRaWANPayload(rxpk, gid, GetLogger(t, "Convert DataRouterReq"))

		// Check
		CheckErrors(t, wantErr, err)
	}

	// --------------------

	{
		Desc(t, "Test toLoRaWANPayload with join-request mtype")

		// Build
		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		payload := lorawan.NewPHYPayload(false)
		payload.MHDR.MType = lorawan.JoinRequest
		payload.MHDR.Major = lorawan.LoRaWANR1
		payload.MIC = [4]byte{1, 2, 3, 4}
		joinpayload := &lorawan.JoinRequestPayload{
			AppEUI:   [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:   [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
			DevNonce: [2]byte{1, 2},
		}
		payload.MACPayload = joinpayload
		data, err := payload.MarshalBinary()
		FatalUnless(t, err)
		rxpk := semtech.RXPK{
			Codr: pointer.String("4/5"),
			Freq: pointer.Float32(867.345),
			Data: pointer.String(base64.RawStdEncoding.EncodeToString(data)),
		}

		// Expectations
		var wantErr *string
		var wantReq = &core.JoinRouterReq{
			GatewayID: gid,
			AppEUI:    joinpayload.AppEUI[:],
			DevEUI:    joinpayload.DevEUI[:],
			DevNonce:  joinpayload.DevNonce[:],
			MIC:       payload.MIC[:],
			Metadata: &core.Metadata{
				CodingRate: *rxpk.Codr,
				Frequency:  *rxpk.Freq,
			},
		}

		// Operate
		req, err := toLoRaWANPayload(rxpk, gid, GetLogger(t, "Convert DataRouterReq"))

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantReq, req, "Join Router Requests")
	}
}
