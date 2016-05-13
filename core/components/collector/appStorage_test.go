// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collector

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func createStorage() AppStorage {
	storage, err := ConnectRedis("localhost:6379", 0)
	if err != nil {
		panic(err)
	}
	return storage
}

func TestConnect(t *testing.T) {
	a := New(t)

	c, err := ConnectRedis("localhost:6379", 0)
	a.So(err, ShouldBeNil)
	defer c.Close()

	_, err = ConnectRedis("", 0)
	a.So(err, ShouldNotBeNil)
}

func TestAccessSetKey(t *testing.T) {
	a := New(t)

	eui, _ := types.ParseAppEUI("8000000000000001")
	key := "key"

	storage := createStorage()
	defer storage.Close()
	defer storage.Reset()

	err := storage.SetAccessKey(eui, key)
	a.So(err, ShouldBeNil)

	fetchedKey, err := storage.GetAccessKey(eui)
	a.So(err, ShouldBeNil)
	a.So(fetchedKey, ShouldEqual, key)
}

func TestList(t *testing.T) {
	a := New(t)

	eui1, _ := types.ParseAppEUI("8000000000000001")
	eui2, _ := types.ParseAppEUI("8000000000000002")

	storage := createStorage()
	defer storage.Close()
	defer storage.Reset()

	err := storage.Add(eui1)
	a.So(err, ShouldBeNil)
	err = storage.Add(eui2)
	a.So(err, ShouldBeNil)
	apps, err := storage.List()
	a.So(err, ShouldBeNil)
	a.So(apps, ShouldHaveLength, 2)

	err = storage.Remove(eui1)
	a.So(err, ShouldBeNil)
	apps, err = storage.List()
	a.So(err, ShouldBeNil)
	a.So(apps, ShouldHaveLength, 1)
}
