// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/storage"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
)

type countingStore struct {
	store  application.Store
	counts map[string]int
}

func newCountingStore(store application.Store) *countingStore {
	return &countingStore{
		store: store,
	}
}

func (s *countingStore) inc(name string) {
	val, ok := s.counts[name]
	if !ok {
		val = 0
	}
	s.counts[name] = val + 1
}

func (s *countingStore) count(name string) int {
	val, ok := s.counts[name]
	if !ok {
		val = 0
	}
	return val
}

func (s *countingStore) Count() (int, error) {
	return s.store.Count()
}

func (s *countingStore) List(opts *storage.ListOptions) ([]*application.Application, error) {
	s.inc("list")
	return s.store.List(opts)
}

func (s *countingStore) Get(appID string) (*application.Application, error) {
	s.inc("get")
	return s.store.Get(appID)
}

func (s *countingStore) Set(app *application.Application, fields ...string) error {
	s.inc("set")
	return s.store.Set(app, fields...)
}

func (s *countingStore) Delete(appID string) error {
	s.inc("delete")
	return s.store.Delete(appID)
}

func TestDryUplinkFieldsCustom(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewRedisApplicationStore(GetRedisClient(), "handler-test-dry-uplink"))
	h := &handler{
		applications: store,
	}
	m := &handlerManager{handler: h}

	dryUplinkMessage := &pb.DryUplinkMessage{
		Payload: []byte{11, 22, 33},
		App: &pb.Application{
			AppId:         "DryUplinkFields",
			PayloadFormat: "custom",
			Decoder: `function Decoder (bytes) {
				console.log("hi", 11)
				return { length: bytes.length }}`,
			Converter: `function Converter (obj) {
				console.log("foo")
				return obj
			}`,
			Validator: `function Validator (bytes) { return true; }`,
		},
	}

	res, err := m.DryUplink(context.TODO(), dryUplinkMessage)
	a.So(err, ShouldBeNil)

	a.So(res.Payload, ShouldResemble, dryUplinkMessage.Payload)
	a.So(res.Fields, ShouldEqual, `{"length":3}`)
	a.So(res.Valid, ShouldBeTrue)
	a.So(res.Logs, ShouldResemble, []*pb.LogEntry{
		&pb.LogEntry{
			Function: "Decoder",
			Fields:   []string{`"hi"`, "11"},
		},
		&pb.LogEntry{
			Function: "Converter",
			Fields:   []string{`"foo"`},
		},
	})

	// make sure no calls to app store were made
	a.So(store.count("list"), ShouldEqual, 0)
	a.So(store.count("get"), ShouldEqual, 0)
	a.So(store.count("set"), ShouldEqual, 0)
	a.So(store.count("delete"), ShouldEqual, 0)
}

func TestDryUplinkFieldsCayenneLPP(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewRedisApplicationStore(GetRedisClient(), "handler-test-dry-uplink"))
	h := &handler{
		applications: store,
	}
	m := &handlerManager{handler: h}

	dryUplinkMessage := &pb.DryUplinkMessage{
		Payload: []byte{7, 103, 0, 245},
		App: &pb.Application{
			AppId:         "DryUplinkFields",
			PayloadFormat: "cayennelpp",
		},
	}

	res, err := m.DryUplink(context.TODO(), dryUplinkMessage)
	a.So(err, ShouldBeNil)

	a.So(res.Payload, ShouldResemble, dryUplinkMessage.Payload)
	a.So(res.Fields, ShouldEqual, `{"temperature_7":24.5}`)
	a.So(res.Valid, ShouldBeTrue)
}

func TestDryUplinkEmptyApp(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewRedisApplicationStore(GetRedisClient(), "handler-test-dry-uplink"))
	h := &handler{
		applications: store,
	}
	m := &handlerManager{handler: h}

	dryUplinkMessage := &pb.DryUplinkMessage{
		Payload: []byte{11, 22, 33},
	}

	res, err := m.DryUplink(context.TODO(), dryUplinkMessage)
	a.So(err, ShouldBeNil)

	a.So(res.Payload, ShouldResemble, dryUplinkMessage.Payload)
	a.So(res.Fields, ShouldEqual, "")
	a.So(res.Valid, ShouldBeTrue)

	// make sure no calls to app store were made
	a.So(store.count("list"), ShouldEqual, 0)
	a.So(store.count("get"), ShouldEqual, 0)
	a.So(store.count("set"), ShouldEqual, 0)
	a.So(store.count("delete"), ShouldEqual, 0)
}

func TestDryDownlinkFieldsCustom(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewRedisApplicationStore(GetRedisClient(), "handler-test-dry-downlink"))
	h := &handler{
		applications: store,
	}
	m := &handlerManager{handler: h}

	msg := &pb.DryDownlinkMessage{
		Fields: `{ "foo": [ 1, 2, 3 ] }`,
		App: &pb.Application{
			PayloadFormat: "custom",
			Encoder: `
				function Encoder (fields) {
					console.log("hello", { foo: 33 })
					return fields.foo
				}`,
		},
	}

	res, err := m.DryDownlink(context.TODO(), msg)
	a.So(err, ShouldBeNil)

	a.So(res.Payload, ShouldResemble, []byte{1, 2, 3})
	a.So(res.Logs, ShouldResemble, []*pb.LogEntry{
		&pb.LogEntry{
			Function: "Encoder",
			Fields:   []string{`"hello"`, `{"foo":33}`},
		},
	})

	// make sure no calls to app store were made
	a.So(store.count("list"), ShouldEqual, 0)
	a.So(store.count("get"), ShouldEqual, 0)
	a.So(store.count("set"), ShouldEqual, 0)
	a.So(store.count("delete"), ShouldEqual, 0)
}

func TestDryDownlinkFieldsCayenneLPP(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewRedisApplicationStore(GetRedisClient(), "handler-test-dry-downlink"))
	h := &handler{
		applications: store,
	}
	m := &handlerManager{handler: h}

	msg := &pb.DryDownlinkMessage{
		Fields: `{ "accelerometer_1": { "x": -0.5, "y": 0.2, "z": 0 } }`,
		App: &pb.Application{
			PayloadFormat: "cayennelpp",
		},
	}

	res, err := m.DryDownlink(context.TODO(), msg)
	a.So(err, ShouldBeNil)

	a.So(res.Payload, ShouldResemble, []byte{1, 113, 254, 12, 0, 200, 0, 0})
}

func TestDryDownlinkPayload(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewRedisApplicationStore(GetRedisClient(), "handler-test-dry-downlink"))
	h := &handler{
		applications: store,
	}
	m := &handlerManager{handler: h}

	msg := &pb.DryDownlinkMessage{
		Payload: []byte{0x1, 0x2, 0x3},
		App: &pb.Application{
			Encoder: `function (fields) { return fields.foo }`,
		},
	}

	res, err := m.DryDownlink(context.TODO(), msg)
	a.So(err, ShouldBeNil)

	a.So(res.Payload, ShouldResemble, []byte{0x1, 0x2, 0x3})
	a.So(res.Logs, ShouldResemble, []*pb.LogEntry(nil))

	// make sure no calls to app store were made
	a.So(store.count("list"), ShouldEqual, 0)
	a.So(store.count("get"), ShouldEqual, 0)
	a.So(store.count("set"), ShouldEqual, 0)
	a.So(store.count("delete"), ShouldEqual, 0)
}

func TestDryDownlinkEmptyApp(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewRedisApplicationStore(GetRedisClient(), "handler-test-dry-downlink"))
	h := &handler{
		applications: store,
	}
	m := &handlerManager{handler: h}

	msg := &pb.DryDownlinkMessage{
		Fields: `{ "foo": [ 1, 2, 3 ] }`,
	}

	_, err := m.DryDownlink(context.TODO(), msg)
	a.So(err, ShouldNotBeNil)

	// make sure no calls to app store were made
	a.So(store.count("list"), ShouldEqual, 0)
	a.So(store.count("get"), ShouldEqual, 0)
	a.So(store.count("set"), ShouldEqual, 0)
	a.So(store.count("delete"), ShouldEqual, 0)
}

func TestLogs(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewRedisApplicationStore(GetRedisClient(), "handler-test-dry-downlink"))
	h := &handler{
		applications: store,
	}
	m := &handlerManager{handler: h}

	msg := &pb.DryDownlinkMessage{
		Fields: `{ "foo": [ 1, 2, 3 ] }`,
		App: &pb.Application{
			PayloadFormat: "custom",
			Encoder: `
				function Encoder (fields) {
					console.log("foo", 1, "bar", new Date(0))
					console.log(1, { baz: 10, baa: "foo", bal: { "bar": 10 }})
					return fields.foo
				}`,
		},
	}

	res, err := m.DryDownlink(context.TODO(), msg)
	a.So(err, ShouldBeNil)
	a.So(res.Logs, ShouldResemble, []*pb.LogEntry{
		&pb.LogEntry{
			Function: "Encoder",
			Fields:   []string{`"foo"`, "1", `"bar"`, `"1970-01-01T00:00:00.000Z"`},
		},
		&pb.LogEntry{
			Function: "Encoder",
			Fields:   []string{"1", `{"baa":"foo","bal":{"bar":10},"baz":10}`},
		},
	})
}
