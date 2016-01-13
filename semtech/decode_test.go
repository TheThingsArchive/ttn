// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"bytes"
	"github.com/thethingsnetwork/ttn/utils/pointer"
	"reflect"
	"testing"
	"time"
)

// ------------------------------------------------------------
// ------------------------- Unmarshal (raw []byte) (Packet, error)
// ------------------------------------------------------------

// Unmarshal() with valid raw data and no payload (PUSH_ACK)
func TestUnmarshal1(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PUSH_ACK}
	packet, err := Unmarshal(raw)

	if err != nil {
		t.Errorf("Failed to parse with error: %#v", err)
	}

	if packet.Version != VERSION {
		t.Errorf("Invalid parsed version: %x", packet.Version)
	}

	if !bytes.Equal([]byte{0x14, 0x14}, packet.Token) {
		t.Errorf("Invalid parsed token: %x", packet.Token)
	}

	if packet.Identifier != PUSH_ACK {
		t.Errorf("Invalid parsed identifier: %x", packet.Identifier)
	}

	if packet.Payload != nil {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
	}

	if packet.GatewayId != nil {
		t.Errorf("Invalid parsed gateway id: % x", packet.GatewayId)
	}
}

// Unmarshal() with valid raw data and stat payload
func TestUnmarshal2(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PUSH_DATA}
	gatewayId := []byte("qwerty1234")[0:8]
	payload := []byte(`{
        "stat": {
            "time":"2014-01-12 08:59:28 GMT",
            "lati":46.24000,
            "long":3.25230,
            "alti":145,
            "rxnb":2,
            "rxok":2,
            "rxfw":2,
            "ackr":100.0,
            "dwnb":2,
            "txnb":2
        }
    }`)
	packet, err := Unmarshal(append(append(raw, gatewayId...), payload...))

	if err != nil {
		t.Errorf("Failed to parse with error: %#v", err)
	}

	if packet.Version != VERSION {
		t.Errorf("Invalid parsed version: %x", packet.Version)
	}

	if !bytes.Equal([]byte{0x14, 0x14}, packet.Token) {
		t.Errorf("Invalid parsed token: %x", packet.Token)
	}

	if !bytes.Equal(gatewayId, packet.GatewayId) {
		t.Errorf("Invalid parsed gatewayId: % x", packet.GatewayId)
	}

	if packet.Identifier != PUSH_DATA {
		t.Errorf("Invalid parsed identifier: %x", packet.Identifier)
	}

	if packet.Payload == nil {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
		return
	}

	if !bytes.Equal(payload, packet.Payload.Raw) {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
	}

	if packet.Payload.Stat == nil {
		t.Errorf("Invalid parsed payload Stat: %#v", packet.Payload.Stat)
	}

	statTime, _ := time.Parse(time.RFC3339, "2014-01-12T08:59:28.000Z")

	stat := Stat{
		Time: &statTime,
		Lati: pointer.Float64(46.24000),
		Long: pointer.Float64(3.25230),
		Alti: pointer.Int(145),
		Rxnb: pointer.Uint(2),
		Rxok: pointer.Uint(2),
		Rxfw: pointer.Uint(2),
		Ackr: pointer.Float64(100.0),
		Dwnb: pointer.Uint(2),
		Txnb: pointer.Uint(2),
	}

	if !reflect.DeepEqual(stat, *packet.Payload.Stat) {
		t.Errorf("Invalid parsed payload Stat: %#v", packet.Payload.Stat)
	}

}

// Unmarshal() with valid raw data and rxpk payloads
func TestUnmarshal3(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PUSH_DATA}
	gatewayId := []byte("qwerty1234")[0:8]
	payload := []byte(`{
        "rxpk":[
            {
                "chan":2,
                "codr":"4/6",
                "data":"-DS4CGaDCdG+48eJNM3Vai-zDpsR71Pn9CPA9uCON84",
                "datr":"SF7BW125",
                "freq":866.349812,
                "lsnr":5.1,
                "modu":"LORA",
                "rfch":0,
                "rssi":-35,
                "size":32,
                "stat":1,
                "time":"2013-03-31T16:21:17.528002Z",
                "tmst":3512348611
            },{
                "chan":9,
                "data":"VEVTVF9QQUNLRVRfMTIzNA==",
                "datr":50000,
                "freq":869.1,
                "modu":"FSK",
                "rfch":1,
                "rssi":-75,
                "size":16,
                "stat":1,
                "time":"2013-03-31T16:21:17.530974Z",
                "tmst":3512348514
            },{
                "chan":0,
                "codr":"4/7",
                "data":"ysgRl452xNLep9S1NTIg2lomKDxUgn3DJ7DE+b00Ass",
                "datr":"SF10BW125",
                "freq":863.00981,
                "lsnr":5.5,
                "modu":"LORA",
                "rfch":0,
                "rssi":-38,
                "size":32,
                "stat":1,
                "time":"2013-03-31T16:21:17.532038Z",
                "tmst":3316387610
            }
        ]
    }`)

	packet, err := Unmarshal(append(append(raw, gatewayId...), payload...))

	if err != nil {
		t.Errorf("Failed to parse with error: %#v", err)
	}

	if packet.Version != VERSION {
		t.Errorf("Invalid parsed version: %x", packet.Version)
	}

	if !bytes.Equal([]byte{0x14, 0x14}, packet.Token) {
		t.Errorf("Invalid parsed token: %x", packet.Token)
	}

	if packet.Identifier != PUSH_DATA {
		t.Errorf("Invalid parsed identifier: %x", packet.Identifier)
	}

	if !bytes.Equal(gatewayId, packet.GatewayId) {
		t.Errorf("Invalid parsed gatewayId: % x", packet.GatewayId)
	}

	if packet.Payload == nil {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
		return
	}

	if !bytes.Equal(payload, packet.Payload.Raw) {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
	}

	if packet.Payload.RXPK == nil || len(packet.Payload.RXPK) != 3 {
		t.Errorf("Invalid parsed payload RXPK: %#v", packet.Payload.RXPK)
	}

	rxpkTime, _ := time.Parse(time.RFC3339, "2013-03-31T16:21:17.530974Z")

	rxpk := RXPK{
		Chan: pointer.Uint(9),
		Data: pointer.String("VEVTVF9QQUNLRVRfMTIzNA=="),
		Datr: pointer.String("50000"),
		Freq: pointer.Float64(869.1),
		Modu: pointer.String("FSK"),
		Rfch: pointer.Uint(1),
		Rssi: pointer.Int(-75),
		Size: pointer.Uint(16),
		Stat: pointer.Int(1),
		Time: &rxpkTime,
		Tmst: pointer.Uint(3512348514),
	}

	if !reflect.DeepEqual(rxpk, (packet.Payload.RXPK)[1]) {
		t.Errorf("Invalid parsed payload RXPK: %#v", packet.Payload.RXPK)
	}
}

// Unmarshal() with valid raw data and rxpk payloads + stat
func TestUnmarshal4(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PUSH_DATA}
	gatewayId := []byte("qwerty1234")[0:8]
	payload := []byte(`{
        "rxpk":[
            {
                "chan":2,
                "codr":"4/6",
                "data":"-DS4CGaDCdG+48eJNM3Vai-zDpsR71Pn9CPA9uCON84",
                "datr":"SF7BW125",
                "freq":866.349812,
                "lsnr":5.1,
                "modu":"LORA",
                "rfch":0,
                "rssi":-35,
                "size":32,
                "stat":1,
                "time":"2013-03-31T16:21:17.528002Z",
                "tmst":3512348611
            },{
                "chan":9,
                "data":"VEVTVF9QQUNLRVRfMTIzNA==",
                "datr":50000,
                "freq":869.1,
                "modu":"FSK",
                "rfch":1,
                "rssi":-75,
                "size":16,
                "stat":1,
                "time":"2013-03-31T16:21:17.530974Z",
                "tmst":3512348514
            },{
                "chan":0,
                "codr":"4/7",
                "data":"ysgRl452xNLep9S1NTIg2lomKDxUgn3DJ7DE+b00Ass",
                "datr":"SF10BW125",
                "freq":863.00981,
                "lsnr":5.5,
                "modu":"LORA",
                "rfch":0,
                "rssi":-38,
                "size":32,
                "stat":1,
                "time":"2013-03-31T16:21:17.532038Z",
                "tmst":3316387610
            }
        ],
        "stat": {
            "time":"2014-01-12 08:59:28 GMT",
            "lati":46.24000,
            "long":3.25230,
            "alti":145,
            "rxnb":2,
            "rxok":2,
            "rxfw":2,
            "ackr":100.0,
            "dwnb":2,
            "txnb":2
        }
    }`)

	packet, err := Unmarshal(append(append(raw, gatewayId...), payload...))

	if err != nil {
		t.Errorf("Failed to parse with error: %#v", err)
	}

	if packet.Version != VERSION {
		t.Errorf("Invalid parsed version: %x", packet.Version)
	}

	if !bytes.Equal([]byte{0x14, 0x14}, packet.Token) {
		t.Errorf("Invalid parsed token: %x", packet.Token)
	}

	if packet.Identifier != PUSH_DATA {
		t.Errorf("Invalid parsed identifier: %x", packet.Identifier)
	}

	if packet.Payload == nil {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
		return
	}

	if !bytes.Equal(payload, packet.Payload.Raw) {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
	}

	if packet.Payload.RXPK == nil || len(packet.Payload.RXPK) != 3 {
		t.Errorf("Invalid parsed payload RXPK: %#v", packet.Payload.RXPK)
	}

	if packet.Payload.Stat == nil {
		t.Errorf("Invalid parsed payload Stat: %#v", packet.Payload.Stat)
	}

	rxpkTime, _ := time.Parse(time.RFC3339, "2013-03-31T16:21:17.530974Z")

	rxpk := RXPK{
		Chan: pointer.Uint(9),
		Data: pointer.String("VEVTVF9QQUNLRVRfMTIzNA=="),
		Datr: pointer.String("50000"),
		Freq: pointer.Float64(869.1),
		Modu: pointer.String("FSK"),
		Rfch: pointer.Uint(1),
		Rssi: pointer.Int(-75),
		Size: pointer.Uint(16),
		Stat: pointer.Int(1),
		Time: &rxpkTime,
		Tmst: pointer.Uint(3512348514),
	}

	if !reflect.DeepEqual(rxpk, (packet.Payload.RXPK)[1]) {
		t.Errorf("Invalid parsed payload RXPK: %#v", packet.Payload.RXPK)
	}

	statTime, _ := time.Parse(time.RFC3339, "2014-01-12T08:59:28.000Z")

	stat := Stat{
		Time: &statTime,
		Lati: pointer.Float64(46.24000),
		Long: pointer.Float64(3.25230),
		Alti: pointer.Int(145),
		Rxnb: pointer.Uint(2),
		Rxok: pointer.Uint(2),
		Rxfw: pointer.Uint(2),
		Ackr: pointer.Float64(100.0),
		Dwnb: pointer.Uint(2),
		Txnb: pointer.Uint(2),
	}

	if !reflect.DeepEqual(stat, *packet.Payload.Stat) {
		t.Errorf("Invalid parsed payload Stat: %#v", packet.Payload.Stat)
	}
}

// Unmarshal() with valid raw data and txpk payload
func TestUnmarshal5(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PULL_RESP}

	payload := []byte(`{
        "txpk":{
            "imme":true,
            "freq":864.123456,
            "rfch":0,
            "powe":14,
            "modu":"LORA",
            "datr":"SF11BW125",
            "codr":"4/6",
            "ipol":false,
            "size":32,
            "data":"H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v"
        }
    }`)

	packet, err := Unmarshal(append(raw, payload...))

	if err != nil {
		t.Errorf("Failed to parse with error: %#v", err)
	}

	if packet.Version != VERSION {
		t.Errorf("Invalid parsed version: %x", packet.Version)
	}

	if !bytes.Equal([]byte{0x14, 0x14}, packet.Token) {
		t.Errorf("Invalid parsed token: %x", packet.Token)
	}

	if packet.Identifier != PULL_RESP {
		t.Errorf("Invalid parsed identifier: %x", packet.Identifier)
	}

	if packet.Payload == nil {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
		return
	}

	if !bytes.Equal(payload, packet.Payload.Raw) {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
	}

	if packet.Payload.TXPK == nil {
		t.Errorf("Invalid parsed payload TXPK: %#v", packet.Payload.TXPK)
	}

	txpk := TXPK{
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
	}

	if !reflect.DeepEqual(txpk, *packet.Payload.TXPK) {
		t.Errorf("Invalid parsed payload TXPK: %#v", packet.Payload.TXPK)
	}
}

// Unmarshal() with an invalid version number
func TestUnmarshal6(t *testing.T) {
	raw := []byte{0x00, 0x14, 0x14, PUSH_ACK}
	_, err := Unmarshal(raw)

	if err == nil {
		t.Errorf("Successfully parsed an incorrect version number")
	}
}

// Unmarshal() with an invalid raw message
func TestUnmarshal7(t *testing.T) {
	raw1 := []byte{VERSION}
	var raw2 []byte
	_, err1 := Unmarshal(raw1)
	_, err2 := Unmarshal(raw2)

	if err1 == nil {
		t.Errorf("Successfully parsed an raw message")
	}

	if err2 == nil {
		t.Errorf("Successfully parsed a nil raw message")
	}
}

// Unmarshal() with an invalid identifier
func TestUnmarshal8(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, 0xFF}
	_, err := Unmarshal(raw)

	if err == nil {
		t.Errorf("Successfully parsed an incorrect identifier")
	}
}

// Unmarshal() with an invalid payload
func TestUnmarshal9(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PUSH_DATA}
	gatewayId := []byte("qwerty1234")[0:8]
	payload := []byte(`wrong`)
	_, err := Unmarshal(append(append(raw, gatewayId...), payload...))
	if err == nil {
		t.Errorf("Successfully parsed an incorrect payload")
	}
}

// Unmarshal() with an invalid date
func TestUnmarshal10(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PUSH_DATA}
	gatewayId := []byte("qwerty1234")[0:8]
	payload := []byte(`{
        "stat": {
            "time":"null",
            "lati":46.24000,
            "long":3.25230,
            "alti":145,
            "rxnb":2,
            "rxok":2,
            "rxfw":2,
            "ackr":100.0,
            "dwnb":2,
            "txnb":2
        }
    }`)
	_, err := Unmarshal(append(append(raw, gatewayId...), payload...))
	if err == nil {
		t.Errorf("Successfully parsed an incorrect payload time")
	}
}

// Unmarshal() with valid raw data but a useless payload
func TestUnmarshal11(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PUSH_ACK}
	payload := []byte(`{
        "stat": {
            "time":"2014-01-12 08:59:28 GMT",
            "lati":46.24000,
            "long":3.25230,
            "alti":145,
            "rxnb":2,
            "rxok":2,
            "rxfw":2,
            "ackr":100.0,
            "dwnb":2,
            "txnb":2
        }
    }`)
	packet, err := Unmarshal(append(raw, payload...))

	if err != nil {
		t.Errorf("Failed to parse a valid PUSH_ACK packet")
	}

	if packet.Payload != nil {
		t.Errorf("Parsed payload on a PUSH_ACK packet")
	}
}

// Unmarshal() with valid raw data but a useless payload
func TestUnmarshal12(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PULL_ACK}
	payload := []byte(`{
        "stat": {
            "time":"2014-01-12 08:59:28 GMT",
            "lati":46.24000,
            "long":3.25230,
            "alti":145,
            "rxnb":2,
            "rxok":2,
            "rxfw":2,
            "ackr":100.0,
            "dwnb":2,
            "txnb":2
        }
    }`)
	packet, err := Unmarshal(append(raw, payload...))

	if err != nil {
		t.Errorf("Failed to parse a valid PULL_ACK packet")
	}

	if packet.Payload != nil {
		t.Errorf("Parsed payload on a PULL_ACK packet")
	}
}

// Unmarshal() with valid raw data but a useless payload
func TestUnmarshal13(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PULL_DATA}
	gatewayId := []byte("qwerty1234")[0:8]
	payload := []byte(`{
        "stat": {
            "time":"2014-01-12 08:59:28 GMT",
            "lati":46.24000,
            "long":3.25230,
            "alti":145,
            "rxnb":2,
            "rxok":2,
            "rxfw":2,
            "ackr":100.0,
            "dwnb":2,
            "txnb":2
        }
    }`)
	packet, err := Unmarshal(append(append(raw, gatewayId...), payload...))

	if err != nil {
		t.Errorf("Failed to parse a valid PULL_DATA packet")
	}

	if packet.Payload != nil {
		t.Errorf("Parsed payload on a PULL_DATA packet")
	}
}

// Unmarshal() with valid raw data and no payload (PULL_ACK)
func TestUnmarshal14(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PULL_ACK}
	packet, err := Unmarshal(raw)

	if err != nil {
		t.Errorf("Failed to parse with error: %#v", err)
	}

	if packet.Version != VERSION {
		t.Errorf("Invalid parsed version: %x", packet.Version)
	}

	if !bytes.Equal([]byte{0x14, 0x14}, packet.Token) {
		t.Errorf("Invalid parsed token: %x", packet.Token)
	}

	if packet.Identifier != PULL_ACK {
		t.Errorf("Invalid parsed identifier: %x", packet.Identifier)
	}

	if packet.Payload != nil {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
	}

	if packet.GatewayId != nil {
		t.Errorf("Invalid parsed gateway id: % x", packet.GatewayId)
	}
}

// Unmarshal() with valid raw data and no payload (PULL_DATA)
func TestUnmarshal15(t *testing.T) {
	raw := []byte{VERSION, 0x14, 0x14, PULL_DATA}
	gatewayId := []byte("qwerty1234")[0:8]
	packet, err := Unmarshal(append(raw, gatewayId...))

	if err != nil {
		t.Errorf("Failed to parse with error: %#v", err)
	}

	if packet.Version != VERSION {
		t.Errorf("Invalid parsed version: %x", packet.Version)
	}

	if !bytes.Equal([]byte{0x14, 0x14}, packet.Token) {
		t.Errorf("Invalid parsed token: %x", packet.Token)
	}

	if packet.Identifier != PULL_DATA {
		t.Errorf("Invalid parsed identifier: %x", packet.Identifier)
	}

	if packet.Payload != nil {
		t.Errorf("Invalid parsed payload: % x", packet.Payload)
	}

	if !bytes.Equal(gatewayId, packet.GatewayId) {
		t.Errorf("Invalid parsed gateway id: % x", packet.GatewayId)
	}
}
