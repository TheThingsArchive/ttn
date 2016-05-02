// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collector

import (
	"testing"

	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core/collection"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func createStorage() AppStorage {
	storage, err := ConnectRedisAppStorage(&redis.Options{Addr: "localhost:6379"})
	if err != nil {
		panic(err)
	}
	return storage
}

func TestConnect(t *testing.T) {
	a := New(t)

	c, err := ConnectRedisAppStorage(&redis.Options{Addr: "localhost:6379"})
	a.So(err, ShouldBeNil)
	defer c.Close()

	_, err = ConnectRedisAppStorage(&redis.Options{Addr: ""})
	a.So(err, ShouldNotBeNil)
}

func TestSetKey(t *testing.T) {
	a := New(t)

	eui, _ := types.ParseAppEUI("8000000000000001")
	key := "key"

	storage := createStorage()
	defer storage.Close()
	defer storage.Reset()

	err := storage.SetKey(eui, key)
	a.So(err, ShouldBeNil)

	app, err := storage.Get(eui)
	a.So(err, ShouldBeNil)
	a.So(app.EUI, ShouldEqual, eui)
	a.So(app.Key, ShouldEqual, key)
}

func TestGetAll(t *testing.T) {
	a := New(t)

	eui1, _ := types.ParseAppEUI("8000000000000001")
	eui2, _ := types.ParseAppEUI("8000000000000002")

	storage := createStorage()
	defer storage.Close()
	defer storage.Reset()

	err := storage.SetKey(eui1, "key1")
	a.So(err, ShouldBeNil)
	err = storage.SetKey(eui2, "key2")
	a.So(err, ShouldBeNil)

	apps, err := storage.GetAll()
	a.So(err, ShouldBeNil)
	a.So(apps, ShouldHaveLength, 2)
}

func TestSetFunctions(t *testing.T) {
	a := New(t)

	eui, _ := types.ParseAppEUI("8000000000000001")
	functions := &collection.Functions{
		Decoder:   `function (payload) { return { size: payload.length; } }`,
		Converter: `function (data) { return data; }`,
		Validator: `function (data) { return data.size % 2 == 0; }`,
	}

	storage := createStorage()
	defer storage.Close()
	defer storage.Reset()

	err := storage.SetFunctions(eui, functions)
	a.So(err, ShouldBeNil)

	app, err := storage.Get(eui)
	a.So(err, ShouldBeNil)
	a.So(app.EUI, ShouldEqual, eui)
	a.So(app.Decoder, ShouldEqual, functions.Decoder)
	a.So(app.Converter, ShouldEqual, functions.Converter)
	a.So(app.Validator, ShouldEqual, functions.Validator)
}
