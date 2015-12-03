// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/thethingsnetwork/core/utils/pointer"
	"io/ioutil"
	"testing"
	"time"
)

// -----------------------------------------------------------------
// ------------------------- Marshal (packet *Packet) ([]byte, error)
// -----------------------------------------------------------------
// ---------- PUSH_DATA
func checkMarshalPUSH_DATA(packet *Packet, payload []byte) error {
	raw, err := Marshal(packet)

	if err != nil {
		return err
	}

	if len(raw) < 12 {
		return errors.New(fmt.Sprintf("Invalid raw sequence length: %d", len(raw)))
	}

	if raw[0] != packet.Version {
		return errors.New(fmt.Sprintf("Invalid raw version: %x", raw[0]))
	}

	if !bytes.Equal(raw[1:3], packet.Token) {
		return errors.New(fmt.Sprintf("Invalid raw token: %x", raw[1:3]))
	}

	if raw[3] != packet.Identifier {
		return errors.New(fmt.Sprintf("Invalid raw identifier: %x", raw[3]))
	}

	if !bytes.Equal(raw[4:12], packet.GatewayId) {
		return errors.New(fmt.Sprintf("Invalid raw gatewayId: % x", raw[4:12]))
	}

	if packet.Payload != nil && !bytes.Equal(raw[12:], payload) {
		return errors.New(fmt.Sprintf("Invalid raw payload: % x", raw[12:]))
	}

	return err
}

// Marshal a basic push_data packet with Stat payload
func TestMarshalPUSH_DATA1(t *testing.T) {
	time1, err := time.Parse(time.RFC3339, "2014-01-12T08:59:28Z")

	// {
	//     "stat": {
	//         "ackr": 100,
	//         "alti": 145,
	//         "dwnb": 2,
	//         "lati": 46.24,
	//         "long": 3.2523,
	//         "rxfw": 2,
	//         "rxnb": 2,
	//         "rxok": 2,
	//         "time": "2014-01-12 08:59:28 GMT",
	//         "txnb": 2
	//     }
	// }
	payload, err := ioutil.ReadFile("./test_data/marshal_stat")

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PUSH_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload: &Payload{
			Stat: &Stat{
				Ackr: pointer.Float64(100.0),
				Alti: pointer.Int(145),
				Long: pointer.Float64(3.25230),
				Rxok: pointer.Uint(2),
				Rxfw: pointer.Uint(2),
				Rxnb: pointer.Uint(2),
				Lati: pointer.Float64(46.24),
				Dwnb: pointer.Uint(2),
				Txnb: pointer.Uint(2),
				Time: &time1,
			},
		},
	}

	if err = checkMarshalPUSH_DATA(packet, payload); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Mashal a push_data packet with RXPK payload
func TestMarshalPUSH_DATA2(t *testing.T) {
	time1, err := time.Parse(time.RFC3339Nano, "2013-03-31T16:21:17.528002Z")
	time2, err := time.Parse(time.RFC3339Nano, "2013-03-31T16:21:17.530974Z")

	//{
	//    "rxpk": [
	//        {
	//            "chan": 2,
	//            "codr": "4/6",
	//            "data": "-DS4CGaDCdG+48eJNM3Vai-zDpsR71Pn9CPA9uCON84",
	//            "datr": "SF7BW125",
	//            "freq": 866.349812,
	//            "lsnr": 5.1,
	//            "modu": "LORA",
	//            "rfch": 0,
	//            "rssi": -35,
	//            "size": 32,
	//            "stat": 1,
	//            "time": "2013-03-31T16:21:17.528002Z",
	//            "tmst": 3512348611
	//        },
	//        {
	//            "chan": 9,
	//            "data": "VEVTVF9QQUNLRVRfMTIzNA==",
	//            "datr": 50000,
	//            "freq": 869.1,
	//            "modu": "FSK",
	//            "rfch": 1,
	//            "rssi": -75,
	//            "size": 16,
	//            "stat": 1,
	//            "time": "2013-03-31T16:21:17.530974Z",
	//            "tmst": 3512348514
	//        }
	//    ]
	//}
	payload, err := ioutil.ReadFile("./test_data/marshal_rxpk")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PUSH_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload: &Payload{
			RXPK: &[]RXPK{
				RXPK{
					Chan: pointer.Uint(2),
					Codr: pointer.String("4/6"),
					Data: pointer.String("-DS4CGaDCdG+48eJNM3Vai-zDpsR71Pn9CPA9uCON84"),
					Datr: pointer.String("SF7BW125"),
					Freq: pointer.Float64(866.349812),
					Lsnr: pointer.Float64(5.1),
					Modu: pointer.String("LORA"),
					Rfch: pointer.Uint(0),
					Rssi: pointer.Int(-35),
					Size: pointer.Uint(32),
					Stat: pointer.Int(1),
					Time: &time1,
					Tmst: pointer.Uint(3512348611),
				},
				RXPK{
					Chan: pointer.Uint(9),
					Data: pointer.String("VEVTVF9QQUNLRVRfMTIzNA=="),
					Datr: pointer.String("50000"),
					Freq: pointer.Float64(869.1),
					Modu: pointer.String("FSK"),
					Rfch: pointer.Uint(1),
					Rssi: pointer.Int(-75),
					Size: pointer.Uint(16),
					Stat: pointer.Int(1),
					Time: &time2,
					Tmst: pointer.Uint(3512348514),
				},
			},
		},
	}

	if err = checkMarshalPUSH_DATA(packet, payload); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Mashal a push_data packet with RXPK payload and Stat
func TestMarshalPUSH_DATA3(t *testing.T) {
	time1, err := time.Parse(time.RFC3339, "2014-01-12T08:59:28Z")
	time2, err := time.Parse(time.RFC3339Nano, "2013-03-31T16:21:17.528002Z")
	time3, err := time.Parse(time.RFC3339Nano, "2013-03-31T16:21:17.530974Z")

	// {
	//     "rxpk": [
	//         {
	//             "chan": 2,
	//             "codr": "4/6",
	//             "data": "-DS4CGaDCdG+48eJNM3Vai-zDpsR71Pn9CPA9uCON84",
	//             "datr": "SF7BW125",
	//             "freq": 866.349812,
	//             "lsnr": 5.1,
	//             "modu": "LORA",
	//             "rfch": 0,
	//             "rssi": -35,
	//             "size": 32,
	//             "stat": 1,
	//             "time": "2013-03-31T16:21:17.528002Z",
	//             "tmst": 3512348611
	//         },
	//         {
	//             "chan": 9,
	//             "data": "VEVTVF9QQUNLRVRfMTIzNA==",
	//             "datr": 50000,
	//             "freq": 869.1,
	//             "modu": "FSK",
	//             "rfch": 1,
	//             "rssi": -75,
	//             "size": 16,
	//             "stat": 1,
	//             "time": "2013-03-31T16:21:17.530974Z",
	//             "tmst": 3512348514
	//         }
	//     ],
	//     "stat": {
	//         "ackr": 100,
	//         "alti": 145,
	//         "dwnb": 2,
	//         "lati": 46.24,
	//         "long": 3.2523,
	//         "rxfw": 2,
	//         "rxnb": 2,
	//         "rxok": 2,
	//         "time": "2014-01-12 08:59:28 GMT",
	//         "txnb": 2
	//     }
	// }
	payload, err := ioutil.ReadFile("./test_data/marshal_rxpk_stat")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PUSH_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload: &Payload{
			Stat: &Stat{
				Ackr: pointer.Float64(100.0),
				Alti: pointer.Int(145),
				Long: pointer.Float64(3.25230),
				Rxok: pointer.Uint(2),
				Rxfw: pointer.Uint(2),
				Rxnb: pointer.Uint(2),
				Lati: pointer.Float64(46.24),
				Dwnb: pointer.Uint(2),
				Txnb: pointer.Uint(2),
				Time: &time1,
			},
			RXPK: &[]RXPK{
				RXPK{
					Time: &time2,
					Tmst: pointer.Uint(3512348611),
					Chan: pointer.Uint(2),
					Rfch: pointer.Uint(0),
					Freq: pointer.Float64(866.349812),
					Stat: pointer.Int(1),
					Modu: pointer.String("LORA"),
					Datr: pointer.String("SF7BW125"),
					Codr: pointer.String("4/6"),
					Rssi: pointer.Int(-35),
					Lsnr: pointer.Float64(5.1),
					Size: pointer.Uint(32),
					Data: pointer.String("-DS4CGaDCdG+48eJNM3Vai-zDpsR71Pn9CPA9uCON84"),
				},
				RXPK{
					Chan: pointer.Uint(9),
					Data: pointer.String("VEVTVF9QQUNLRVRfMTIzNA=="),
					Datr: pointer.String("50000"),
					Freq: pointer.Float64(869.1),
					Modu: pointer.String("FSK"),
					Rfch: pointer.Uint(1),
					Rssi: pointer.Int(-75),
					Size: pointer.Uint(16),
					Stat: pointer.Int(1),
					Time: &time3,
					Tmst: pointer.Uint(3512348514),
				},
			},
		},
	}

	if err = checkMarshalPUSH_DATA(packet, payload); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal with an invalid GatewayId (too short)
func TestMarshalPUSH_DATA4(t *testing.T) {
	time1, err := time.Parse(time.RFC3339, "2014-01-12T08:59:28Z")

	// {"stat":{"ackr":100,"alti":145,"dwnb":2,"lati":46.24,"long":3.2523,"rxfw":2,"rxnb":2,"rxok":2,"time":"2014-01-12 08:59:28 GMT","txnb":2}}
	payload, err := ioutil.ReadFile("./test_data/marshal_stat")

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PUSH_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, // Invalid
		Payload: &Payload{
			Stat: &Stat{
				Ackr: pointer.Float64(100.0),
				Alti: pointer.Int(145),
				Long: pointer.Float64(3.25230),
				Rxok: pointer.Uint(2),
				Rxfw: pointer.Uint(2),
				Rxnb: pointer.Uint(2),
				Lati: pointer.Float64(46.24),
				Dwnb: pointer.Uint(2),
				Txnb: pointer.Uint(2),
				Time: &time1,
			},
		},
	}

	if err = checkMarshalPUSH_DATA(packet, payload); err == nil {
		t.Errorf("Successfully mashalled a invalid packet")
	}
}

// Marshal with an invalid GatewayId (too long)
func TestMarshalPUSH_DATA5(t *testing.T) {
	time1, err := time.Parse(time.RFC3339, "2014-01-12T08:59:28Z")

	// {"stat":{"ackr":100,"alti":145,"dwnb":2,"lati":46.24,"long":3.2523,"rxfw":2,"rxnb":2,"rxok":2,"time":"2014-01-12 08:59:28 GMT","txnb":2}}
	payload, err := ioutil.ReadFile("./test_data/marshal_stat")

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PUSH_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, // Invalid
		Payload: &Payload{
			Stat: &Stat{
				Ackr: pointer.Float64(100.0),
				Alti: pointer.Int(145),
				Long: pointer.Float64(3.25230),
				Rxok: pointer.Uint(2),
				Rxfw: pointer.Uint(2),
				Rxnb: pointer.Uint(2),
				Lati: pointer.Float64(46.24),
				Dwnb: pointer.Uint(2),
				Txnb: pointer.Uint(2),
				Time: &time1,
			},
		},
	}

	if err = checkMarshalPUSH_DATA(packet, payload); err == nil {
		t.Errorf("Successfully mashalled a invalid packet")
	}
}

// Marshal with an invalid TokenId (too short)
func TestMarshalPUSH_DATA6(t *testing.T) {
	time1, err := time.Parse(time.RFC3339, "2014-01-12T08:59:28Z")

	// {"stat":{"ackr":100,"alti":145,"dwnb":2,"lati":46.24,"long":3.2523,"rxfw":2,"rxnb":2,"rxok":2,"time":"2014-01-12 08:59:28 GMT","txnb":2}}
	payload, err := ioutil.ReadFile("./test_data/marshal_stat")

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA},
		Identifier: PUSH_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, // Invalid
		Payload: &Payload{
			Stat: &Stat{
				Ackr: pointer.Float64(100.0),
				Alti: pointer.Int(145),
				Long: pointer.Float64(3.25230),
				Rxok: pointer.Uint(2),
				Rxfw: pointer.Uint(2),
				Rxnb: pointer.Uint(2),
				Lati: pointer.Float64(46.24),
				Dwnb: pointer.Uint(2),
				Txnb: pointer.Uint(2),
				Time: &time1,
			},
		},
	}

	if err = checkMarshalPUSH_DATA(packet, payload); err == nil {
		t.Errorf("Successfully mashalled a invalid packet")
	}
}

// Marshal with an invalid TokenId (too long)
func TestMarshalPUSH_DATA7(t *testing.T) {
	time1, err := time.Parse(time.RFC3339, "2014-01-12T08:59:28Z")

	// {"stat":{"ackr":100,"alti":145,"dwnb":2,"lati":46.24,"long":3.2523,"rxfw":2,"rxnb":2,"rxok":2,"time":"2014-01-12 08:59:28 GMT","txnb":2}}
	payload, err := ioutil.ReadFile("./test_data/marshal_stat")

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14, 0x28},
		Identifier: PUSH_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, // Invalid
		Payload: &Payload{
			Stat: &Stat{
				Ackr: pointer.Float64(100.0),
				Alti: pointer.Int(145),
				Long: pointer.Float64(3.25230),
				Rxok: pointer.Uint(2),
				Rxfw: pointer.Uint(2),
				Rxnb: pointer.Uint(2),
				Lati: pointer.Float64(46.24),
				Dwnb: pointer.Uint(2),
				Txnb: pointer.Uint(2),
				Time: &time1,
			},
		},
	}

	if err = checkMarshalPUSH_DATA(packet, payload); err == nil {
		t.Errorf("Successfully mashalled a invalid packet")
	}
}

// ---------- PUSH_ACK
func checkMarshalACK(packet *Packet) error {
	raw, err := Marshal(packet)

	if err != nil {
		return err
	}

	if len(raw) != 4 {
		return errors.New(fmt.Sprintf("Invalid raw sequence length: %d", len(raw)))
	}

	if raw[0] != packet.Version {
		return errors.New(fmt.Sprintf("Invalid raw version: %x", raw[0]))
	}

	if !bytes.Equal(raw[1:3], packet.Token) {
		return errors.New(fmt.Sprintf("Invalid raw token: %x", raw[1:3]))
	}

	if raw[3] != packet.Identifier {
		return errors.New(fmt.Sprintf("Invalid raw identifier: %x", raw[3]))
	}

	return err
}

// Marshal a basic push_ack packet
func TestMarshalPUSH_ACK1(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PUSH_ACK,
		GatewayId:  nil,
		Payload:    nil,
	}
	if err := checkMarshalACK(packet); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal a push_ack packet with extra useless gatewayId
func TestMarshalPUSH_ACK2(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PUSH_ACK,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload:    nil,
	}
	if err := checkMarshalACK(packet); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal a push_ack packet with extra useless Payload
func TestMarshalPUSH_ACK3(t *testing.T) {
	payload := &Payload{
		Stat: &Stat{
			Rxfw: pointer.Uint(14),
			Rxnb: pointer.Uint(14),
			Rxok: pointer.Uint(14),
		},
	}
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PUSH_ACK,
		GatewayId:  nil,
		Payload:    payload,
	}
	if err := checkMarshalACK(packet); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal a push_ack with extra useless gatewayId and payload
func TestMarshalPUSH_ACK4(t *testing.T) {
	payload := &Payload{
		Stat: &Stat{
			Rxfw: pointer.Uint(14),
			Rxnb: pointer.Uint(14),
			Rxok: pointer.Uint(14),
		},
	}
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PUSH_ACK,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload:    payload,
	}
	if err := checkMarshalACK(packet); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal a push_ack with an invalid token (too short)
func TestMarshalPUSH_ACK5(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA},
		Identifier: PUSH_ACK,
		GatewayId:  nil,
		Payload:    nil,
	}
	_, err := Marshal(packet)
	if err == nil {
		t.Errorf("Successfully marshalled an invalid packet")
	}
}

// Marshal a push_ack with an invalid token (too long)
func TestMarshalPUSH_ACK6(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0x9A, 0x7A, 0x7E},
		Identifier: PUSH_ACK,
		GatewayId:  nil,
		Payload:    nil,
	}
	_, err := Marshal(packet)
	if err == nil {
		t.Errorf("Successfully marshalled an invalid packet")
	}
}

// ---------- PULL_DATA
func checkMarshalPULL_DATA(packet *Packet) error {
	raw, err := Marshal(packet)

	if err != nil {
		return err
	}

	if len(raw) != 12 {
		return errors.New(fmt.Sprintf("Invalid raw sequence length: %d", len(raw)))
	}

	if raw[0] != packet.Version {
		return errors.New(fmt.Sprintf("Invalid raw version: %x", raw[0]))
	}

	if !bytes.Equal(raw[1:3], packet.Token) {
		return errors.New(fmt.Sprintf("Invalid raw token: %x", raw[1:3]))
	}

	if raw[3] != packet.Identifier {
		return errors.New(fmt.Sprintf("Invalid raw identifier: %x", raw[3]))
	}

	if !bytes.Equal(raw[4:12], packet.GatewayId) {
		return errors.New(fmt.Sprintf("Invalid raw gatewayId: % x", raw[4:12]))
	}

	return err
}

// Marshal a basic pull_data packet
func TestMarshalPULL_DATA1(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload:    nil,
	}
	if err := checkMarshalPULL_DATA(packet); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal a pull_data packet with an extra useless Payload
func TestMarshalPULL_DATA2(t *testing.T) {
	payload := &Payload{
		Stat: &Stat{
			Rxfw: pointer.Uint(14),
			Rxnb: pointer.Uint(14),
			Rxok: pointer.Uint(14),
		},
	}
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload:    payload,
	}
	if err := checkMarshalPULL_DATA(packet); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal a pull_data packet with an invalid token (too short)
func TestMarshalPULL_DATA3(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA},
		Identifier: PULL_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload:    nil,
	}
	if err := checkMarshalPULL_DATA(packet); err == nil {
		t.Errorf("Successfully marshalled a packet with an invalid token")
	}
}

// Marshal a pull_data packet with an invalid token (too long)
func TestMarshalPULL_DATA4(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14, 0x42},
		Identifier: PULL_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload:    nil,
	}
	if err := checkMarshalPULL_DATA(packet); err == nil {
		t.Errorf("Successfully marshalled a packet with an invalid token")
	}
}

// Marshal a pull_data packet with an invalid gatewayId (too short)
func TestMarshalPULL_DATA5(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload:    nil,
	}
	if err := checkMarshalPULL_DATA(packet); err == nil {
		t.Errorf("Successfully marshalled a packet with an invalid gatewayId")
	}
}

// Marshal a pull_data packet with an invalid gatewayId (too long)
func TestMarshalPULL_DATA6(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_DATA,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload:    nil,
	}
	if err := checkMarshalPULL_DATA(packet); err == nil {
		t.Errorf("Successfully marshalled a packet with an invalid gatewayId")
	}
}

// ---------- PULL_ACK
// Marshal a basic pull_ack packet
func TestMarshalPULL_ACK1(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_ACK,
		GatewayId:  nil,
		Payload:    nil,
	}
	if err := checkMarshalACK(packet); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal a pull_ack packet with extra useless gatewayId
func TestMarshalPULL_ACK2(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_ACK,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload:    nil,
	}
	if err := checkMarshalACK(packet); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal a pull_ack packet with extra useless Payload
func TestMarshalPULL_ACK3(t *testing.T) {
	payload := &Payload{
		Stat: &Stat{
			Rxfw: pointer.Uint(14),
			Rxnb: pointer.Uint(14),
			Rxok: pointer.Uint(14),
		},
	}
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_ACK,
		GatewayId:  nil,
		Payload:    payload,
	}
	if err := checkMarshalACK(packet); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal a pull_ack with extra useless gatewayId and payload
func TestMarshalPULL_ACK4(t *testing.T) {
	payload := &Payload{
		Stat: &Stat{
			Rxfw: pointer.Uint(14),
			Rxnb: pointer.Uint(14),
			Rxok: pointer.Uint(14),
		},
	}
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_ACK,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload:    payload,
	}
	if err := checkMarshalACK(packet); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal a pull_ack with an invalid token (too short)
func TestMarshalPULL_ACK5(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA},
		Identifier: PULL_ACK,
		GatewayId:  nil,
		Payload:    nil,
	}
	_, err := Marshal(packet)
	if err == nil {
		t.Errorf("Successfully marshalled an invalid packet")
	}
}

// Marshal a pull_ack with an invalid token (too long)
func TestMarshalPULL_ACK6(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0x9A, 0x7A, 0x7E},
		Identifier: PULL_ACK,
		GatewayId:  nil,
		Payload:    nil,
	}
	_, err := Marshal(packet)
	if err == nil {
		t.Errorf("Successfully marshalled an invalid packet")
	}
}

// ---------- PULL_RESP
func checkMarshalPULL_RESP(packet *Packet, payload []byte) error {
	raw, err := Marshal(packet)

	if err != nil {
		return err
	}

	if len(raw) < 4 {
		return errors.New(fmt.Sprintf("Invalid raw sequence length: %d", len(raw)))
	}

	if raw[0] != packet.Version {
		return errors.New(fmt.Sprintf("Invalid raw version: %x", raw[0]))
	}

	if !bytes.Equal(raw[1:3], packet.Token) {
		return errors.New(fmt.Sprintf("Invalid raw token: %x", raw[1:3]))
	}

	if raw[3] != packet.Identifier {
		return errors.New(fmt.Sprintf("Invalid raw identifier: %x", raw[3]))
	}

	if packet.Payload != nil && !bytes.Equal(raw[4:], payload) {
		return errors.New(fmt.Sprintf("Invalid raw payload: % x", raw[4:]))
	}

	return err
}

// Marshal() for a basic PULL_RESP packet with no payload
func TestMarshallPULL_RESP1(t *testing.T) {
	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_RESP,
		GatewayId:  nil,
		Payload:    nil,
	}

	if err := checkMarshalPULL_RESP(packet, make([]byte, 0)); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal() for a basic PULL_RESP packet with RXPK payload
func TestMarshallPULL_RESP2(t *testing.T) {

	// {
	//     "txpk": {
	//         "codr": "4/6",
	//         "data": "H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v",
	//         "datr": "SF11BW125",
	//         "freq": 864.123456,
	//         "imme": true,
	//         "ipol": false,
	//         "modu": "LORA",
	//         "powe": 14,
	//         "rfch": 0,
	//         "size": 32
	//     }
	// }
	payload, err := ioutil.ReadFile("./test_data/marshal_txpk")

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_RESP,
		GatewayId:  nil,
		Payload: &Payload{
			TXPK: &TXPK{
				Imme: pointer.Bool(true),
				Freq: pointer.Float64(864.123456),
				Rfch: pointer.Uint(0),
				Powe: pointer.Uint(14),
				Modu: pointer.String("LORA"),
				Datr: pointer.String("SF11BW125"),
				Codr: pointer.String("4/6"),
				Ipol: pointer.Bool(false),
				Size: pointer.Uint(32),
				Data: pointer.String("H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v"),
			},
		},
	}

	if err = checkMarshalPULL_RESP(packet, payload); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal() for a basic PULL_RESP packet with RXPK payload and useless gatewayId
func TestMarshallPULL_RESP3(t *testing.T) {
	//{"txpk":{"codr":"4/6","data":"H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v","datr":"SF11BW125","freq":864.123456,"imme":true,"ipol":false,"modu":"LORA","powe":14,"rfch":0,"size":32}}
	payload, err := ioutil.ReadFile("./test_data/marshal_txpk")

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14},
		Identifier: PULL_RESP,
		GatewayId:  []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Payload: &Payload{
			TXPK: &TXPK{
				Imme: pointer.Bool(true),
				Freq: pointer.Float64(864.123456),
				Rfch: pointer.Uint(0),
				Powe: pointer.Uint(14),
				Modu: pointer.String("LORA"),
				Datr: pointer.String("SF11BW125"),
				Codr: pointer.String("4/6"),
				Ipol: pointer.Bool(false),
				Size: pointer.Uint(32),
				Data: pointer.String("H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v"),
			},
		},
	}

	if err = checkMarshalPULL_RESP(packet, payload); err != nil {
		t.Errorf("Failed to marshal packet: %v", err)
	}
}

// Marshal() for a PULL_RESP packet with an invalid token (too short)
func TestMarshallPULL_RESP4(t *testing.T) {
	//{"txpk":{"codr":"4/6","data":"H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v","datr":"SF11BW125","freq":864.123456,"imme":true,"ipol":false,"modu":"LORA","powe":14,"rfch":0,"size":32}}
	payload, err := ioutil.ReadFile("./test_data/marshal_txpk")

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA},
		Identifier: PULL_RESP,
		GatewayId:  nil,
		Payload: &Payload{
			TXPK: &TXPK{
				Imme: pointer.Bool(true),
				Freq: pointer.Float64(864.123456),
				Rfch: pointer.Uint(0),
				Powe: pointer.Uint(14),
				Modu: pointer.String("LORA"),
				Datr: pointer.String("SF11BW125"),
				Codr: pointer.String("4/6"),
				Ipol: pointer.Bool(false),
				Size: pointer.Uint(32),
				Data: pointer.String("H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v"),
			},
		},
	}

	if err = checkMarshalPULL_RESP(packet, payload); err == nil {
		t.Errorf("Successfully marshalled a packet with an invalid token")
	}
}

// Marshal() for a PULL_RESP packet with an invalid token (too long)
func TestMarshallPULL_RESP5(t *testing.T) {
	//{"txpk":{"codr":"4/6","data":"H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v","datr":"SF11BW125","freq":864.123456,"imme":true,"ipol":false,"modu":"LORA","powe":14,"rfch":0,"size":32}}
	payload, err := ioutil.ReadFile("./test_data/marshal_txpk")

	packet := &Packet{
		Version:    VERSION,
		Token:      []byte{0xAA, 0x14, 0x42},
		Identifier: PULL_RESP,
		GatewayId:  nil,
		Payload: &Payload{
			TXPK: &TXPK{
				Imme: pointer.Bool(true),
				Freq: pointer.Float64(864.123456),
				Rfch: pointer.Uint(0),
				Powe: pointer.Uint(14),
				Modu: pointer.String("LORA"),
				Datr: pointer.String("SF11BW125"),
				Codr: pointer.String("4/6"),
				Ipol: pointer.Bool(false),
				Size: pointer.Uint(32),
				Data: pointer.String("H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v"),
			},
		},
	}

	if err = checkMarshalPULL_RESP(packet, payload); err == nil {
		t.Errorf("Successfully marshalled a packet with an invalid token")
	}
}
