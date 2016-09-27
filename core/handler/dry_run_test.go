// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	"golang.org/x/net/context"

	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	. "github.com/smartystreets/assertions"
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

func (s *countingStore) Count(name string) int {
	val, ok := s.counts[name]
	if !ok {
		val = 0
	}
	return val
}

func (s *countingStore) List() ([]*application.Application, error) {
	s.inc("list")
	return s.store.List()
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

func TestDryUplinkFields(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewApplicationStore())
	h := &handler{
		applications: store,
	}
	m := &handlerManager{handler: h}

	dryUplinkMessage := &pb.DryUplinkMessage{
		Payload: []byte{11, 22, 33},
		App: &pb.Application{
			Decoder:   `function (bytes) { return { length: bytes.length }}`,
			Converter: `function (obj) { return obj }`,
			Validator: `function (bytes) { return true; }`,
		},
	}

	res, err := m.DryUplink(context.TODO(), dryUplinkMessage)
	a.So(err, ShouldBeNil)

	a.So(res.Payload, ShouldResemble, dryUplinkMessage.Payload)
	a.So(res.Fields, ShouldEqual, `{"length":3}`)
	a.So(res.Valid, ShouldBeTrue)

	// make sure no calls to app store were made
	a.So(store.Count("list"), ShouldEqual, 0)
	a.So(store.Count("get"), ShouldEqual, 0)
	a.So(store.Count("set"), ShouldEqual, 0)
	a.So(store.Count("delete"), ShouldEqual, 0)
}

func TestDryUplinkEmptyApp(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewApplicationStore())
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
	a.So(store.Count("list"), ShouldEqual, 0)
	a.So(store.Count("get"), ShouldEqual, 0)
	a.So(store.Count("set"), ShouldEqual, 0)
	a.So(store.Count("delete"), ShouldEqual, 0)
}

func TestDryDownlinkFields(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewApplicationStore())
	h := &handler{
		applications: store,
	}
	m := &handlerManager{handler: h}

	msg := &pb.DryDownlinkMessage{
		Fields: `{ "foo": [ 1, 2, 3 ] }`,
		App: &pb.Application{
			Encoder: `function (fields) { return fields.foo }`,
		},
	}

	res, err := m.DryDownlink(context.TODO(), msg)
	a.So(err, ShouldBeNil)

	a.So(res.Payload, ShouldResemble, []byte{1, 2, 3})

	// make sure no calls to app store were made
	a.So(store.Count("list"), ShouldEqual, 0)
	a.So(store.Count("get"), ShouldEqual, 0)
	a.So(store.Count("set"), ShouldEqual, 0)
	a.So(store.Count("delete"), ShouldEqual, 0)
}

func TestDryDownlinkPayload(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewApplicationStore())
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

	// make sure no calls to app store were made
	a.So(store.Count("list"), ShouldEqual, 0)
	a.So(store.Count("get"), ShouldEqual, 0)
	a.So(store.Count("set"), ShouldEqual, 0)
	a.So(store.Count("delete"), ShouldEqual, 0)
}

func TestDryDownlinkEmptyApp(t *testing.T) {
	a := New(t)

	store := newCountingStore(application.NewApplicationStore())
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
	a.So(store.Count("list"), ShouldEqual, 0)
	a.So(store.Count("get"), ShouldEqual, 0)
	a.So(store.Count("set"), ShouldEqual, 0)
	a.So(store.Count("delete"), ShouldEqual, 0)
}
