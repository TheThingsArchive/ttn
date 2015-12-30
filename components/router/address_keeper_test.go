// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"testing"
	"time"
)

func genDevAddr() semtech.DeviceAddress {
	return semtech.DeviceAddress(fmt.Sprintf("DeviceAddress%d", time.Now()))
}

func genBroAddr() core.BrokerAddress {
	return core.BrokerAddress(fmt.Sprintf("BrokerAddress%d", time.Now()))
}

func TestAddressKeeper(t *testing.T) {
	Convey("Local Address Keeper", t, func() {
		Convey("NewLocalDB", func() {
			Convey("NewLocalDB: valid", func() {
				localDB, err := NewLocalDB(time.Hour * 100)
				So(err, ShouldBeNil)
				So(localDB, ShouldNotBeNil)
			})

			Convey("NewLocalDB: invalid", func() {
				localDB, err := NewLocalDB(0)
				So(err, ShouldNotBeNil)
				So(localDB, ShouldBeNil)
			})
		})

		Convey("Store & Lookup", func() {
			Convey("Store then Lookup same", func() {
				localDB, err := NewLocalDB(time.Hour)
				if err != nil {
					panic(err)
				}

				devAddr := genDevAddr()
				broAddr := genBroAddr()

				err = localDB.store(devAddr, broAddr)
				So(err, ShouldBeNil)

				broAddrs, err := localDB.lookup(devAddr)
				So(err, ShouldBeNil)
				So(broAddrs, ShouldResemble, []core.BrokerAddress{broAddr})

				devAddr = genDevAddr()
				broAddr2 := genBroAddr()
				err = localDB.store(devAddr, broAddr, broAddr2)
				So(err, ShouldBeNil)

				broAddrs, err = localDB.lookup(devAddr)
				So(err, ShouldBeNil)
				So(broAddrs, ShouldResemble, []core.BrokerAddress{broAddr, broAddr2})
			})

			Convey("Invalid lookups", func() {
				localDB, err := NewLocalDB(time.Millisecond)
				if err != nil {
					panic(err)
				}

				devAddr := genDevAddr()
				broAddr := genBroAddr()

				err = localDB.store(devAddr, broAddr)
				So(err, ShouldBeNil)

				time.Sleep(10 * time.Millisecond)

				broAddrs, err := localDB.lookup(devAddr)
				So(broAddrs, ShouldBeNil)
				So(err, ShouldNotBeNil)

				broAddrs, err = localDB.lookup(genDevAddr())
				So(err, ShouldNotBeNil)
				So(broAddrs, ShouldBeNil)
			})

			Convey("Store existing", func() {
				localDB, err := NewLocalDB(time.Hour)
				if err != nil {
					panic(err)
				}

				devAddr := genDevAddr()
				broAddr := genBroAddr()

				err = localDB.store(devAddr, broAddr)
				So(err, ShouldBeNil)
				err = localDB.store(devAddr, broAddr)
				So(err, ShouldNotBeNil)
			})
		})

	})
}
