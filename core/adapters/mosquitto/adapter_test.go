// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mosquitto

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

/**

- Setup a client
- Create topic on the client and register to them
- Publish to the registration topic and see if nextRegistrationIsTriggered
- Publish to registered topics and see if next() trigger things
- Publish to unregistered topics and make sure next() isn't triggered
- Send() to a given topic and see if client received

*/

type publicationShape struct {
	AppEUI  string
	DevEUI  string
	Topic   string
	Content interface{}
}

type packetShape struct {
	DevAddr lorawan.DevAddr
	Data    string
}

func TestNext(t *testing.T) {
	devices := []PersonnalizedActivation{
		{
			DevAddr: lorawan.DevAddr([4]byte{0, 0, 0, 1}),
			NwkSKey: lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			AppSKey: lorawan.AES128Key([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}),
		},
		{
			DevAddr: lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			NwkSKey: lorawan.AES128Key([16]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}),
			AppSKey: lorawan.AES128Key([16]byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2}),
		},
	}

	applications := []lorawan.EUI64{
		lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
		lorawan.EUI64([8]byte{1, 1, 2, 2, 3, 3, 4, 4}),
	}

	tests := []struct {
		Desc          string
		Registrations []publicationShape
		Publication   publicationShape
		WantPacket    packetShape
		WantError     *string
	}{
		{
			Desc: "Register #0 | Publish #0 -> #0",
			Registrations: []publicationShape{
				{
					AppEUI:  hex.EncodeToString(applications[0][:]),
					DevEUI:  "personnalized",
					Topic:   TOPIC_ACTIVATIONS,
					Content: devices[0],
				},
			},
			Publication: publicationShape{
				AppEUI:  hex.EncodeToString(applications[0][:]),
				DevEUI:  fmt.Sprintf("%s%s", hex.EncodeToString([]byte{0, 0, 0, 0}), hex.EncodeToString(devices[0].DevAddr[:])),
				Topic:   TOPIC_UPLINK,
				Content: "Data",
			},
			WantPacket: packetShape{
				DevAddr: devices[0].DevAddr,
				Data:    "Data",
			},
			WantError: nil,
		},
	}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		adapter, mosquitto := genAdapter(t, test.Registrations)

		// Operate
		mosquitto.Publish(test.Publication)
		packet, _, err := adapter.Next()

		// Check
		checkErrors(t, test.WantError, err)
		checkPackets(t, test.WantPacket, packet)
	}
}

// ----- BUILD utilities
type Mosquitto struct {
}

func (m *Mosquitto) Publish(p publicationShape) {

}

func genAdapter(t *testing.T, registrations []publicationShape) (*Adapter, *Mosquitto) {
	return nil, nil
}

// ----- OPERATE utilities

// ----- CHECK utilities
func checkErrors(t *testing.T, want *string, got error) {

}

func checkPackets(t *testing.T, want packetShape, got core.Packet) {

}
