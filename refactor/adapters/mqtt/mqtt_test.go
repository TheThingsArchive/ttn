// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	core "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const brokerUrl = "0.0.0.0:1683"

func TestMQTTSend(t *testing.T) {
	tests := []struct {
		Desc       string          // Test Description
		Packet     []byte          // Handy representation of the packet to send
		Recipients []testRecipient // List of recipient to send

		WantData     []byte  // Expected Data on the recipient
		WantResponse []byte  // Expected Response from the Send method
		WantError    *string // Expected error nature returned by the Send method
	}{
		{
			Desc:   "1 packet | 1 recipient | No response",
			Packet: []byte("TheThingsNetwork"),
			Recipients: []testRecipient{
				{
					Response:  nil,
					TopicUp:   "up1",
					TopicDown: "down1",
				},
			},

			WantData:     []byte("TheThingsNetwork"),
			WantResponse: nil,
			WantError:    pointer.String(string(errors.Behavioural)),
		},
		{
			Desc:   "invalid packet | 1 recipient | No response",
			Packet: nil,
			Recipients: []testRecipient{
				{
					Response:  nil,
					TopicUp:   "up1",
					TopicDown: "down1",
				},
			},

			WantData:     nil,
			WantResponse: nil,
			WantError:    pointer.String(string(errors.Structural)),
		},
		{
			Desc:   "1 packet | 2 recipient | No response",
			Packet: []byte("TheThingsNetwork"),
			Recipients: []testRecipient{
				{
					Response:  nil,
					TopicUp:   "up1",
					TopicDown: "down1",
				},
				{
					Response:  nil,
					TopicUp:   "up2",
					TopicDown: "down2",
				},
			},

			WantData:     []byte("TheThingsNetwork"),
			WantResponse: nil,
			WantError:    pointer.String(string(errors.Behavioural)),
		},
		{
			Desc:   "1 packet | 2 recipients | #1 answer ",
			Packet: []byte("TheThingsNetwork"),
			Recipients: []testRecipient{
				{
					Response:  []byte("IoT Rocks"),
					TopicUp:   "up1",
					TopicDown: "down1",
				},
				{
					Response:  nil,
					TopicUp:   "up2",
					TopicDown: "down2",
				},
			},

			WantData:     []byte("TheThingsNetwork"),
			WantResponse: []byte("IoT Rocks"),
			WantError:    nil,
		},
		{
			Desc:   "1 packet | 2 recipients | both answers ",
			Packet: []byte("TheThingsNetwork"),
			Recipients: []testRecipient{
				{
					Response:  []byte("IoT Rocks"),
					TopicUp:   "up1",
					TopicDown: "down1",
				},
				{
					Response:  []byte("IoT Rocks"),
					TopicUp:   "up2",
					TopicDown: "down2",
				},
			},

			WantData:     []byte("TheThingsNetwork"),
			WantResponse: nil,
			WantError:    pointer.String(string(errors.Behavioural)),
		},
	}

	for i, test := range tests {
		// Describe
		Desc(t, fmt.Sprintf("#%d: %s", i, test.Desc))

		// Build
		aclient, adapter := createAdapter(t)
		sclient, chresp := createServers(test.Recipients)
		<-time.After(time.Millisecond * 50)

		// Operate
		resp, err := trySend(adapter, test.Packet, test.Recipients)
		var data []byte
		select {
		case data = <-chresp:
		case <-time.After(time.Millisecond * 100):
		}

		// Check
		checkErrors(t, test.WantError, err)
		checkData(t, test.WantData, data)
		checkResponses(t, test.WantResponse, resp)

		// Clean
		aclient.Disconnect(250)
		sclient.Disconnect(250)
		<-time.After(time.Millisecond * 50)
	}
}

func TestMQTTRecipient(t *testing.T) {
	{
		Desc(t, "Marshal / Unmarshal valid recipient")
		rm := NewRecipient("topicup", "topicdown")
		ru := new(mqttRecipient)
		data, err := rm.MarshalBinary()
		if err == nil {
			err = ru.UnmarshalBinary(data)
		}
		checkErrors(t, nil, err)
	}

	{
		Desc(t, "Unmarshal from nil pointer")
		rm := NewRecipient("topicup", "topicdown")
		var ru *mqttRecipient
		data, err := rm.MarshalBinary()
		if err == nil {
			err = ru.UnmarshalBinary(data)
		}
		checkErrors(t, pointer.String(string(errors.Structural)), err)
	}

	{
		Desc(t, "Unmarshal nil data")
		ru := new(mqttRecipient)
		err := ru.UnmarshalBinary(nil)
		checkErrors(t, pointer.String(string(errors.Structural)), err)
	}

	{
		Desc(t, "Unmarshal wrong data")
		ru := new(mqttRecipient)
		err := ru.UnmarshalBinary([]byte{1, 2, 3, 4})
		checkErrors(t, pointer.String(string(errors.Structural)), err)
	}
}

// ----- TYPE utilities
type testRecipient struct {
	Response  []byte
	TopicUp   string
	TopicDown string
}

type testPacket struct {
	payload []byte
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p testPacket) MarshalBinary() ([]byte, error) {
	if p.payload == nil {
		return nil, errors.New(errors.Structural, "Fake error")
	}

	return p.payload, nil
}

// String implements the core.Packet interface
func (p testPacket) String() string {
	return string(p.payload)
}

// ----- BUILD utilities
func createAdapter(t *testing.T) (*MQTT.Client, core.Adapter) {
	client, err := NewClient("testClient", brokerUrl, Tcp)
	if err != nil {
		panic(err)
	}

	adapter := NewAdapter(client, GetLogger(t, "adapter"))
	return client, adapter
}

func createServers(recipients []testRecipient) (*MQTT.Client, chan []byte) {
	client, err := NewClient("FakeServerClient", brokerUrl, Tcp)
	if err != nil {
		panic(err)
	}

	chresp := make(chan []byte, len(recipients))
	for _, r := range recipients {
		go func(r testRecipient) {
			token := client.Subscribe(r.TopicUp, 2, func(client *MQTT.Client, msg MQTT.Message) {
				if r.Response != nil {
					token := client.Publish(r.TopicDown, 2, false, r.Response)
					if token.Wait() && token.Error() != nil {
						panic(token.Error())
					}
				}
				chresp <- msg.Payload()
			})
			if token.Wait() && token.Error() != nil {
				panic(token.Error())
			}
		}(r)
	}
	return client, chresp
}

// ----- OPERATE utilities
func trySend(adapter core.Adapter, packet []byte, recipients []testRecipient) ([]byte, error) {
	// Convert testRecipient to core.Recipient using the mqtt recipient
	var coreRecipients []core.Recipient
	for _, r := range recipients {
		coreRecipients = append(coreRecipients, NewRecipient(r.TopicUp, r.TopicDown))
	}

	// Try send the packet
	chresp := make(chan struct {
		Data  []byte
		Error error
	})
	go func() {
		data, err := adapter.Send(testPacket{packet}, coreRecipients...)
		chresp <- struct {
			Data  []byte
			Error error
		}{data, err}
	}()

	select {
	case resp := <-chresp:
		return resp.Data, resp.Error
	case <-time.After(time.Millisecond * 1250):
		return nil, nil
	}
}

// ----- CHECK utilities
func checkErrors(t *testing.T, want *string, got error) {
	if got == nil {
		if want == nil {
			Ok(t, "Check errors")
			return
		}
		Ko(t, "Expected error to be {%s} but got nothing", *want)
		return
	}

	if want == nil {
		Ko(t, "Expected no error but got {%v}", got)
		return
	}

	if got.(errors.Failure).Nature == errors.Nature(*want) {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be {%s} but got {%v}", *want, got)
}

func checkResponses(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check responses")
		return
	}
	Ko(t, "Received response does not match expectations.\nWant: %s\nGot:  %s", string(want), string(got))
}

func checkData(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check data")
		return
	}
	Ko(t, "Received data does not match expectations.\nWant: %s\nGot:  %s", string(want), string(got))
}
