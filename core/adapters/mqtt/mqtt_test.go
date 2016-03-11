// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"testing"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

const brokerURL = "0.0.0.0:1883"

func TestMQTTSend(t *testing.T) {
	timeout = 100 * time.Millisecond

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
			WantError:    nil,
		},
		{
			Desc:   "1 packet | 1 recipient | No down topic",
			Packet: []byte("TheThingsNetwork"),
			Recipients: []testRecipient{
				{
					Response:  nil,
					TopicUp:   "up1",
					TopicDown: "",
				},
			},

			WantData:     []byte("TheThingsNetwork"),
			WantResponse: nil,
			WantError:    nil,
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
			WantError:    nil,
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
		sclients, chresp := createServers(test.Recipients)
		<-time.After(time.Millisecond * 50)

		// Operate
		resp, err := trySend(adapter, test.Packet, test.Recipients)
		var data []byte
		select {
		case data = <-chresp:
		case <-time.After(time.Millisecond * 100):
		}

		// Check
		CheckErrors(t, test.WantError, err)
		checkData(t, test.WantData, data)
		checkResponses(t, test.WantResponse, resp)

		// Clean
		<-time.After(time.Millisecond * 500)
		aclient.Disconnect(0)
		for _, sclient := range sclients {
			sclient.Disconnect(0)
		}
	}
}

func TestSendErrorCases(t *testing.T) {
	tests := []struct {
		Desc       string          // Test Description
		Packet     []byte          // Handy representation of the packet to send
		Recipients []testRecipient // List of recipient to send
		Client     *MockClient     // A mocked version of the client

		WantData  []byte  // Expected Data on the recipient
		WantError *string // Expected error nature returned by the Send method
	}{
		{
			Desc:   "1 packet | 1 Recipient | Error on publish",
			Packet: []byte("TheThingsNetwork"),
			Client: NewMockClient("Publish"),
			Recipients: []testRecipient{
				{
					Response:  nil,
					TopicUp:   "up",
					TopicDown: "down",
				},
			},

			WantData:  []byte("TheThingsNetwork"),
			WantError: pointer.String(string(errors.Operational)),
		},
		{
			Desc:   "1 packet | 1 Recipient | Error on Subscribe",
			Packet: []byte("TheThingsNetwork"),
			Client: NewMockClient("Subscribe"),
			Recipients: []testRecipient{
				{
					Response:  nil,
					TopicUp:   "up",
					TopicDown: "down",
				},
			},

			WantData:  nil,
			WantError: pointer.String(string(errors.Operational)),
		},
		{
			Desc:   "1 packet | 1 Recipient | Error on Unsubscribe",
			Packet: []byte("TheThingsNetwork"),
			Client: NewMockClient("Unsubscribe"),
			Recipients: []testRecipient{
				{
					Response:  nil,
					TopicUp:   "up",
					TopicDown: "down",
				},
			},

			WantData:  []byte("TheThingsNetwork"),
			WantError: nil,
		},
	}

	for i, test := range tests {
		// Describe
		Desc(t, fmt.Sprintf("#%d: %s", i, test.Desc))

		// Build
		adapter := NewAdapter(test.Client, GetLogger(t, "Adapter"))

		// Operate
		_, err := trySend(adapter, test.Packet, test.Recipients)

		// Check
		CheckErrors(t, test.WantError, err)
		checkData(t, test.WantData, test.Client.InPublish.Payload())
	}
}

func TestOtherMethods(t *testing.T) {
	{
		// Describe
		Desc(t, "Get Recipient | Wrong data")

		// Build
		adapter := NewAdapter(NewMockClient(), GetLogger(t, "Adapter"))

		// Operate
		_, err := adapter.GetRecipient([]byte{})

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
	}

	// --------------------

	{
		// Describe
		Desc(t, "Get Recipient | Valid Recipient")

		// Build
		adapter := NewAdapter(NewMockClient(), GetLogger(t, "Adapter"))
		data, _ := NewRecipient("up", "down").MarshalBinary()

		// Operate
		recipient, err := adapter.GetRecipient(data)

		// Check
		CheckErrors(t, nil, err)
		checkRecipients(t, NewRecipient("up", "down"), recipient)

	}

	// --------------------

	{
		// Describe
		Desc(t, "Send invalid recipients")

		// Build
		adapter := NewAdapter(NewMockClient(), GetLogger(t, "Adapter"))

		// Operate
		_, err := adapter.Send(mocks.NewMockPacket(), mocks.NewMockRecipient())

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
	}

	// --------------------

	{
		// Describe
		Desc(t, "Bind a new handler")

		// Build
		client := NewMockClient()
		adapter := NewAdapter(client, GetLogger(t, "Adapter"))
		handler := NewMockHandler()
		msg := MockMessage{
			topic:   "MessageTopic",
			payload: []byte{1, 2, 3, 4},
		}

		// Operate
		err := adapter.Bind(handler)
		client.InSubscribeCallBack(client, msg)

		// Check
		CheckErrors(t, nil, err)
		checkMessages(t, msg, handler.InMessage)
	}

	// --------------------

	{
		// Describe
		Desc(t, "Bind a new handler | fails to handle")

		// Build
		client := NewMockClient()
		adapter := NewAdapter(client, GetLogger(t, "Adapter"))
		handler := NewMockHandler()
		handler.Failures["Handle"] = errors.New(errors.Operational, "Mock Error")
		msg := MockMessage{
			topic:   "MessageTopic",
			payload: []byte{1, 2, 3, 4},
		}

		// Operate
		err := adapter.Bind(handler)
		client.InSubscribeCallBack(client, msg)

		// Check
		CheckErrors(t, nil, err)
		checkMessages(t, msg, handler.InMessage)
	}

	// --------------------

	{
		// Describe
		Desc(t, "Bind a new handler | fails to subscribe")

		// Build
		client := NewMockClient("Subscribe")
		adapter := NewAdapter(client, GetLogger(t, "Adapter"))
		handler := NewMockHandler()

		// Operate
		err := adapter.Bind(handler)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
	}
}

func TestMQTTRecipient(t *testing.T) {
	{
		Desc(t, "Marshal / Unmarshal valid recipient")
		rm := NewRecipient("topicup", "topicdown")
		ru := new(recipient)
		data, err := rm.MarshalBinary()
		if err == nil {
			err = ru.UnmarshalBinary(data)
		}
		CheckErrors(t, nil, err)
	}

	{
		Desc(t, "Unmarshal from nil pointer")
		rm := NewRecipient("topicup", "topicdown")
		var ru *recipient
		data, err := rm.MarshalBinary()
		if err == nil {
			err = ru.UnmarshalBinary(data)
		}
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
	}

	{
		Desc(t, "Unmarshal nil data")
		ru := new(recipient)
		err := ru.UnmarshalBinary(nil)
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
	}

	{
		Desc(t, "Unmarshal wrong data")
		ru := new(recipient)
		err := ru.UnmarshalBinary([]byte{1, 2, 3, 4})
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
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

// DevEUI implements the core.Packet interface
func (p testPacket) DevEUI() lorawan.EUI64 {
	return lorawan.EUI64{}
}

// ----- BUILD utilities
func createAdapter(t *testing.T) (Client, core.Adapter) {
	client, err := NewClient("testClient", brokerURL, TCP)
	if err != nil {
		panic(err)
	}

	adapter := NewAdapter(client, GetLogger(t, "adapter"))
	return client, adapter
}

func createServers(recipients []testRecipient) ([]Client, chan []byte) {
	var clients []Client
	chresp := make(chan []byte, len(recipients))
	for i, r := range recipients {
		client, err := NewClient(fmt.Sprintf("FakeServerClient%d", i), brokerURL, TCP)
		if err != nil {
			panic(err)
		}
		clients = append(clients, client)

		go func(r testRecipient, client Client) {
			token := client.Subscribe(r.TopicUp, 2, func(client Client, msg MQTT.Message) {
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
		}(r, client)
	}
	return clients, chresp
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
	case <-time.After(timeout + time.Millisecond*100):
		return nil, nil
	}
}

// ----- CHECK utilities
func checkResponses(t *testing.T, want []byte, got []byte) {
	mocks.Check(t, want, got, "Responses")
}

func checkData(t *testing.T, want []byte, got []byte) {
	mocks.Check(t, want, got, "Data")
}

func checkRecipients(t *testing.T, want core.Recipient, got core.Recipient) {
	mocks.Check(t, want, got, "Recipients")
}

func checkMessages(t *testing.T, want MQTT.Message, got MQTT.Message) {
	mocks.Check(t, want, got, "Messages")
}
