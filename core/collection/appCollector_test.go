// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collection

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	ttntesting "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/apex/log"

	. "github.com/smartystreets/assertions"
)

type mockStorage struct {
	entry chan *mockStorageEntry
}

type mockStorageEntry struct {
	devEUI types.DevEUI
	fields map[string]interface{}
}

func createTestCollector(ctx log.Interface, storage DataStorage) AppCollector {
	eui, _ := types.ParseAppEUI("8000000000000001")
	return NewMqttAppCollector(ctx, "localhost:1883", eui, "", storage)
}

func TestStart(t *testing.T) {
	a := New(t)

	ctx := ttntesting.GetLogger(t, "Collection")
	storage := &mockStorage{}
	collector := createTestCollector(ctx, storage)
	a.So(collector, ShouldNotBeNil)

	err := collector.Start()
	defer collector.Stop()
	a.So(err, ShouldBeNil)
}

func TestCollect(t *testing.T) {
	a := New(t)

	ctx := ttntesting.GetLogger(t, "Collection")
	storage := &mockStorage{
		entry: make(chan *mockStorageEntry),
	}
	collector := createTestCollector(ctx, storage)

	err := collector.Start()
	defer collector.Stop()
	a.So(err, ShouldBeNil)

	appEUI, _ := types.ParseAppEUI("8000000000000001")
	devEUI, _ := types.ParseDevEUI("1000000000000001")

	client := mqtt.NewClient(ctx, "collector", "", "", "tcp://localhost:1883")
	err = client.Connect()
	So(err, ShouldBeNil)
	defer client.Disconnect()

	req := core.DataUpAppReq{
		DevEUI:   devEUI.String(),
		FCnt:     0,
		FPort:    1,
		Metadata: []core.AppMetadata{core.AppMetadata{ServerTime: time.Now().Format(time.RFC3339)}},
		Payload:  []byte{0x1, 0x2, 0x3},
		Fields:   map[string]interface{}{"size": 3},
	}
	if token := client.PublishUplink(appEUI, devEUI, req); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	var entry *mockStorageEntry
	select {
	case entry = <-storage.entry:
		break
	case <-time.After(time.Second):
		panic("Timeout")
	}
	a.So(entry, ShouldNotBeNil)
	a.So(entry.devEUI, ShouldResemble, devEUI)
	a.So(entry.fields, ShouldHaveLength, 1)
	a.So(entry.fields["size"], ShouldEqual, 3)
}

func (s *mockStorage) Save(appEUI types.AppEUI, devEUI types.DevEUI, t time.Time, fields map[string]interface{}) error {
	s.entry <- &mockStorageEntry{devEUI, fields}
	return nil
}

func (s *mockStorage) Close() error {
	s.entry <- nil
	return nil
}
