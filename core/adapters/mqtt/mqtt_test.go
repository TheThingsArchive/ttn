// Copyright © 2016 T//e Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
	"golang.org/x/net/context"
)

func TestHandleDataDown(t *testing.T) {
	{
		Desc(t, "Invalid topic :: TTN")

		// Build
		msg := Msg{
			Topic:   "TTN",
			Payload: []byte(`{"payload":"patate"}`),
			Type:    Down,
		}
		var want *core.DataDownHandlerReq

		// Operate
		req, err := handleDataDown(msg)

		// Check
		CheckErrors(t, ErrStructural, err)
		Check(t, want, req, "DataDown Handler Requests")
	}

	// --------------------

	{
		Desc(t, "Invalid topic :: 01/devices/0102030405060708/down")

		// Build
		msg := Msg{
			Topic:   "01/devices/0102030405060708/down",
			Payload: []byte(`{"payload":"patate"}`),
			Type:    Down,
		}
		var want *core.DataDownHandlerReq

		// Operate
		req, err := handleDataDown(msg)

		// Check
		CheckErrors(t, ErrStructural, err)
		Check(t, want, req, "DataDown Handler Requests")
	}

	// --------------------

	{
		Desc(t, "Invalid topic :: 0102030405060708/devices/010203040506/down")

		// Build
		msg := Msg{
			Topic:   "0102030405060708/devices/010203040506/down",
			Payload: []byte(`{"payload":"patate"}`),
			Type:    Down,
		}
		var want *core.DataDownHandlerReq

		// Operate
		req, err := handleDataDown(msg)

		// Check
		CheckErrors(t, ErrStructural, err)
		Check(t, want, req, "DataDown Handler Requests")
	}

	// --------------------

	{
		Desc(t, "Valid topic, invalid Message Pack payload")

		// Build
		msg := Msg{
			Topic:   "0102030405060708/devices/0910111213141516/down",
			Payload: []byte{129, 167, 112, 97, 121, 108, 111, 97, 100, 150, 112, 97, 116, 97, 116, 101},
			Type:    Down,
		}
		var want *core.DataDownHandlerReq

		// Operate
		req, err := handleDataDown(msg)

		// Check
		CheckErrors(t, ErrStructural, err)
		Check(t, want, req, "DataDown Handler Requests")
	}

	// --------------------

	{
		Desc(t, "Valid topic, invalid json")

		// Build
		msg := Msg{
			Topic:   "0102030405060708/devices/0910111213141516/down",
			Payload: []byte(`{"ttn":14}`),
			Type:    Down,
		}
		var want *core.DataDownHandlerReq

		// Operate
		req, err := handleDataDown(msg)

		// Check
		CheckErrors(t, ErrStructural, err)
		Check(t, want, req, "DataDown Handler Requests")
	}

	// --------------------

	{
		Desc(t, "Valid topic, valid JSON payload")

		// Build
		msg := Msg{
			Topic:   "0102030405060708/devices/0910111213141516/down",
			Payload: []byte(`{"payload":[112,97,116,97,116,101]}`),
			Type:    Down,
		}
		want := &core.DataDownHandlerReq{
			Payload: []byte("patate"),
			AppEUI:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			DevEUI:  []byte{0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16},
		}

		// Operate
		req, err := handleDataDown(msg)

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, req, "DataDown Handler Requests")
	}

	// --------------------

	{
		Desc(t, "Valid topic, valid Message Pack payload")

		// Build
		msg := Msg{
			Topic:   "0102030405060708/devices/0910111213141516/down",
			Payload: []byte{129, 167, 112, 97, 121, 108, 111, 97, 100, 196, 6, 112, 97, 116, 97, 116, 101},
			Type:    Down,
		}
		want := &core.DataDownHandlerReq{
			Payload: []byte("patate"),
			AppEUI:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			DevEUI:  []byte{0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16},
		}

		// Operate
		req, err := handleDataDown(msg)

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, req, "DataDown Handler Requests")
	}
}

func TestConsumeMQTTMsg(t *testing.T) {
	{
		Desc(t, "Consume Valid MsgDown")

		// Build
		options := Options{}
		components := Components{
			Client: NewMockClient(),
			Ctx:    GetLogger(t, "MQTT Adapter"),
		}
		adapter := New(components, options)
		handler := mocks.NewHandlerServer()
		chmsg := make(chan Msg)
		adapter.Start(chmsg, handler)

		wantDown := &core.DataDownHandlerReq{
			Payload: []byte("patate"),
			DevEUI:  []byte{8, 7, 6, 5, 4, 3, 2, 1},
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Operate
		chmsg <- Msg{
			Type:    Down,
			Topic:   "0102030405060708/devices/0807060504030201/down",
			Payload: []byte(`{"payload":[112,97,116,97,116,101]}`),
		}

		// Checks
		Check(t, wantDown, handler.InHandleDataDown.Req, "Handler Down Requests")

		// Clean
		close(chmsg)
	}

	// --------------------

	{
		Desc(t, "Consume invalid MsgDown")

		// Build
		options := Options{}
		components := Components{
			Client: NewMockClient(),
			Ctx:    GetLogger(t, "MQTT Adapter"),
		}
		adapter := New(components, options)
		handler := mocks.NewHandlerServer()
		chmsg := make(chan Msg)
		adapter.Start(chmsg, handler)

		var wantDown *core.DataDownHandlerReq

		// Operate
		chmsg <- Msg{
			Type:    Down,
			Topic:   "0102030405060708/devices/08070605040/down",
			Payload: []byte(`{"payload":[112,97,116,97,116,101]}`),
		}

		// Checks
		Check(t, wantDown, handler.InHandleDataDown.Req, "Handler Down Requests")

		// Clean
		close(chmsg)
	}

	// --------------------

	{
		Desc(t, "Consume Invalid Message Type")

		// Build
		options := Options{}
		components := Components{
			Client: NewMockClient(),
			Ctx:    GetLogger(t, "MQTT Adapter"),
		}
		adapter := New(components, options)
		handler := mocks.NewHandlerServer()
		chmsg := make(chan Msg)
		adapter.Start(chmsg, handler)

		var wantDown *core.DataDownHandlerReq

		// Operate
		chmsg <- Msg{
			Type:    14,
			Topic:   "0102030405060708/devices/0807060504030201/down",
			Payload: []byte(`{"payload":[112,97,116,97,116,101]}`),
		}

		// Checks
		Check(t, wantDown, handler.InHandleDataDown.Req, "Handler Down Requests")

		// Clean
		close(chmsg)
	}
}

func TestHandleData(t *testing.T) {
	{
		Desc(t, "Handle Invalid AppReq -> Empty payload")

		// Build
		options := Options{}
		components := Components{
			Client: NewMockClient(),
			Ctx:    GetLogger(t, "MQTT Adapter"),
		}
		adapter := New(components, options)

		// Expectations
		var wantRes *core.DataAppRes
		var wantPub *client.PublishOptions
		var wantErr = ErrStructural

		// Operate
		res, err := adapter.HandleData(
			context.Background(),
			&core.DataAppReq{
				Payload:  nil,
				AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
				DevEUI:   []byte{0, 0, 0, 0, 1, 2, 3, 4},
				Metadata: []*core.Metadata{},
			},
		)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Responses")
		Check(t, wantPub, components.Client.(*MockClient).InPublish.Options, "Publications")
	}

	// --------------------

	{
		Desc(t, "Handle Invalid AppReq -> Invalid AppEUI")

		// Build
		options := Options{}
		components := Components{
			Client: NewMockClient(),
			Ctx:    GetLogger(t, "MQTT Adapter"),
		}
		adapter := New(components, options)

		// Expectations
		var wantRes *core.DataAppRes
		var wantPub *client.PublishOptions
		var wantErr = ErrStructural

		// Operate
		res, err := adapter.HandleData(
			context.Background(),
			&core.DataAppReq{
				Payload:  []byte("patate"),
				AppEUI:   []byte{1, 2, 3, 4, 5, 6},
				DevEUI:   []byte{0, 0, 0, 0, 1, 2, 3, 4},
				Metadata: []*core.Metadata{},
			},
		)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Responses")
		Check(t, wantPub, components.Client.(*MockClient).InPublish.Options, "Publications")
	}

	// --------------------

	{
		Desc(t, "Handle Invalid AppReq -> Invalid DevEUI")

		// Build
		options := Options{}
		components := Components{
			Client: NewMockClient(),
			Ctx:    GetLogger(t, "MQTT Adapter"),
		}
		adapter := New(components, options)

		// Expectations
		var wantRes *core.DataAppRes
		var wantPub *client.PublishOptions
		var wantErr = ErrStructural

		// Operate
		res, err := adapter.HandleData(
			context.Background(),
			&core.DataAppReq{
				Payload:  []byte("patate"),
				AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
				DevEUI:   []byte{0, 0, 0, 0, 1, 2, 3, 4, 5, 6},
				Metadata: []*core.Metadata{},
			},
		)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Responses")
		Check(t, wantPub, components.Client.(*MockClient).InPublish.Options, "Publications")
	}

	// --------------------

	{
		Desc(t, "Handle Invalid AppReq -> No Metadata")

		// Build
		options := Options{}
		components := Components{
			Client: NewMockClient(),
			Ctx:    GetLogger(t, "MQTT Adapter"),
		}
		adapter := New(components, options)

		// Expectations
		var wantRes *core.DataAppRes
		var wantPub *client.PublishOptions
		var wantErr = ErrStructural

		// Operate
		res, err := adapter.HandleData(
			context.Background(),
			&core.DataAppReq{
				Payload:  []byte("patate"),
				AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
				DevEUI:   []byte{0, 0, 0, 0, 1, 2, 3, 4},
				Metadata: nil,
			},
		)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Responses")
		Check(t, wantPub, components.Client.(*MockClient).InPublish.Options, "Publications")
	}

	// --------------------

	{
		Desc(t, "Handle Invalid AppReq -> Nil AppReq")

		// Build
		options := Options{}
		components := Components{
			Client: NewMockClient(),
			Ctx:    GetLogger(t, "MQTT Adapter"),
		}
		adapter := New(components, options)

		// Expectations
		var wantRes *core.DataAppRes
		var wantPub *client.PublishOptions
		var wantErr = ErrStructural

		// Operate
		res, err := adapter.HandleData(
			context.Background(),
			nil,
		)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Responses")
		Check(t, wantPub, components.Client.(*MockClient).InPublish.Options, "Publications")
	}

	// --------------------

	{
		Desc(t, "Handle Valid AppReq, Fail to Publish")

		// Build
		options := Options{}
		components := Components{
			Client: NewMockClient(),
			Ctx:    GetLogger(t, "MQTT Adapter"),
		}
		components.Client.(*MockClient).Failures["Publish"] = fmt.Errorf("Mock Error")
		adapter := New(components, options)
		msg := core.DataUpAppReq{Payload: []byte("patate"), Metadata: []core.AppMetadata{}}
		data, err := msg.MarshalMsg(nil)
		FatalUnless(t, err)

		// Expectations
		var wantRes *core.DataAppRes
		var wantPub = &client.PublishOptions{
			QoS:       mqtt.QoS2,
			Retain:    false,
			TopicName: []byte("0102030405060708/devices/0000000001020304/up"),
			Message:   data,
		}
		var wantErr = ErrOperational

		// Operate
		res, err := adapter.HandleData(
			context.Background(),
			&core.DataAppReq{
				Payload:  []byte("patate"),
				AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
				DevEUI:   []byte{0, 0, 0, 0, 1, 2, 3, 4},
				Metadata: []*core.Metadata{},
			},
		)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Responses")
		Check(t, wantPub, components.Client.(*MockClient).InPublish.Options, "Publications")
	}

	// --------------------

	{
		Desc(t, "Handle Valid AppReq, Publish successful")

		// Build
		options := Options{}
		components := Components{
			Client: NewMockClient(),
			Ctx:    GetLogger(t, "MQTT Adapter"),
		}
		adapter := New(components, options)
		msg := core.DataUpAppReq{Payload: []byte("patate"), Metadata: []core.AppMetadata{}}
		data, err := msg.MarshalMsg(nil)
		FatalUnless(t, err)

		// Expectations
		var wantRes *core.DataAppRes
		var wantPub = &client.PublishOptions{
			QoS:       mqtt.QoS2,
			Retain:    false,
			TopicName: []byte("0102030405060708/devices/0000000001020304/up"),
			Message:   data,
		}
		var wantErr *string

		// Operate
		res, err := adapter.HandleData(
			context.Background(),
			&core.DataAppReq{
				Payload:  []byte("patate"),
				AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
				DevEUI:   []byte{0, 0, 0, 0, 1, 2, 3, 4},
				Metadata: []*core.Metadata{},
			},
		)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Responses")
		Check(t, wantPub, components.Client.(*MockClient).InPublish.Options, "Publications")
	}
}
