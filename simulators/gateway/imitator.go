// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/simulators/node"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/brocaar/lorawan"
)

// MockWithSchedule will generate fake traffic based on a json file referencing packets
//
// The following shape is expected for the JSON file:
// [
//     {
//         "dev_addr": "0102aabb",
//         "payload": "My Data Payload",
//         "metadata": {
//				"rssi": -40,
//				"modu": "LORA",
//				"datr": "4/7",
//				...
//			},
//	   },
//     ....
// ]
//
// The imitator will fire udp packet every given interval of time to a set of router. Routers
// addresses are expected to contains the destination port as well.
func MockWithSchedule(filename string, delay time.Duration, routers ...string) {
	rxpks, err := rxpkFromConf(filename)
	if err != nil {
		panic(err)
	}

	var adapters []io.ReadWriteCloser
	for _, router := range routers {
		addr, err := net.ResolveUDPAddr("udp", router)
		if err != nil {
			panic(err)
		}
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			panic(err)
		}
		adapters = append(adapters, conn)
	}

	log.SetHandler(text.New(os.Stdout))
	log.SetLevel(log.DebugLevel)
	ctx := log.WithFields(log.Fields{"Simulator": "Gateway"})

	fwd, err := NewForwarder([8]byte{1, 2, 3, 4, 5, 6, 7, 8}, ctx, adapters...)
	if err != nil {
		panic(err)
	}

	for {
		for _, rxpk := range rxpks {
			<-time.After(delay)
			if err := fwd.Forward(rxpk); err != nil {
				ctx.WithError(err).WithField("rxpk", rxpk).Warn("failed to forward")
			}
		}
	}
}

func MockRandomly(nodes []node.LiveNode, ctx log.Interface, routers ...string) {
	var adapters []io.ReadWriteCloser
	for _, router := range routers {
		addr, err := net.ResolveUDPAddr("udp", router)
		if err != nil {
			panic(err)
		}
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			panic(err)
		}
		adapters = append(adapters, conn)
	}

	fwd, err := NewForwarder([8]byte{1, 2, 3, 4, 5, 6, 7, 8}, ctx, adapters...)
	if err != nil {
		panic(err)
	}

	messages := make(chan string)

	for _, n := range nodes {
		ctx.Infof("Created node: %s", n.(fmt.Stringer).String())
		go n.Start(messages)
	}

	for {
		rxpks := [8]semtech.RXPK{}
		numPks := rand.Intn(8) + 1

		for i := 0; i < numPks; i++ {
			message := <-messages

			rxpk := semtech.RXPK{
				Rssi: pointer.Int(-20),
				Datr: pointer.String("SF7BW125"),
				Modu: pointer.String("LORA"),
				Data: pointer.String(message),
			}

			rxpks[i] = rxpk
		}

		err := fwd.Forward(rxpks[:numPks]...)
		if err != nil {
			ctx.WithError(err).Warn("failed to forward")
		} else {
			ctx.Debugf("Forwarded %d packets.", numPks)
		}
	}
}

// rxpkFromConf read an input json file and parse it into a list of RXPK packets
func rxpkFromConf(filename string) ([]semtech.RXPK, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var conf []struct {
		DevAddr  string   `json:"dev_addr"`
		Metadata metadata `json:"metadata"`
		Payload  string   `json:"payload"`
	}
	err = json.Unmarshal(content, &conf)
	if err != nil {
		return nil, err
	}
	var rxpks []semtech.RXPK
	for _, c := range conf {
		rxpk := semtech.RXPK(c.Metadata)
		rawAddr, err := hex.DecodeString(c.DevAddr)
		if err != nil {
			return nil, err
		}
		var devAddr lorawan.DevAddr
		copy(devAddr[:], rawAddr)
		rxpk.Data = pointer.String(generateData(c.Payload, devAddr))
		rxpks = append(rxpks, rxpk)
	}

	return rxpks, nil
}

type metadata semtech.RXPK // metadata is just an alias used to mislead the UnmarshalJSON

// UnmarshalJSON implements the json.Unmarshal interface
func (m *metadata) UnmarshalJSON(raw []byte) error {
	if m == nil {
		return fmt.Errorf("Cannot unmarshal in nil metadata")
	}
	payload := new(semtech.Payload)
	rawPayload := append(append([]byte(`{"rxpk":[`), raw...), []byte(`]}`)...)
	err := json.Unmarshal(rawPayload, payload)
	if err != nil {
		return err
	}
	if len(payload.RXPK) < 1 {
		return fmt.Errorf("Unable to interpret raw bytes as valid metadata")
	}
	*m = metadata(payload.RXPK[0])
	return nil
}
