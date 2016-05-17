// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collector

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	ttntesting "github.com/TheThingsNetwork/ttn/utils/testing"

	. "github.com/smartystreets/assertions"
)

type mockStorage struct {
}

func TestCollection(t *testing.T) {
	a := New(t)

	appStorage, err := ConnectRedis("localhost:6379", 0)
	if err != nil {
		panic(err)
	}
	defer appStorage.Close()
	defer appStorage.Reset()

	eui, _ := types.ParseAppEUI("8000000000000001")
	err = appStorage.Add(eui)
	a.So(err, ShouldBeNil)
	err = appStorage.SetAccessKey(eui, "secret")
	a.So(err, ShouldBeNil)

	collector := NewCollector(ttntesting.GetLogger(t, "Collector"), appStorage, "localhost:1883", &mockStorage{}, "localhost:1783")
	collectors, err := collector.Start()
	defer collector.Stop()
	a.So(err, ShouldBeNil)
	a.So(collectors, ShouldHaveLength, 1)

	err = collector.StopApp(eui)
	a.So(err, ShouldBeNil)

	err = collector.StopApp(eui)
	a.So(err, ShouldNotBeNil) // Not found
}

func (s *mockStorage) Save(appEUI types.AppEUI, devEUI types.DevEUI, t time.Time, fields map[string]interface{}) error {
	return nil
}

func (s *mockStorage) Close() error {
	return nil
}
