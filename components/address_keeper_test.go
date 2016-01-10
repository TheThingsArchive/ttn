// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
	"testing"
	"time"
)

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

				devAddr := lorawan.DevAddr([4]byte{1, 2, 3, 4})
				recipient := core.Recipient{Address: "MyAddress", Id: "MyId"}

				err = localDB.store(devAddr, recipient)
				So(err, ShouldBeNil)

				recipients, err := localDB.lookup(devAddr)
				So(err, ShouldBeNil)
				So(recipients, ShouldResemble, []core.Recipient{recipient})

				devAddr = lorawan.DevAddr([4]byte{3, 4, 5, 2})
				recipient2 := core.Recipient{Address: "MyAddress2", Id: "MyId2"}
				err = localDB.store(devAddr, recipient, recipient2)
				So(err, ShouldBeNil)

				recipients, err = localDB.lookup(devAddr)
				So(err, ShouldBeNil)
				So(recipients, ShouldResemble, []core.Recipient{recipient, recipient2})
			})

			Convey("Invalid lookups", func() {
				localDB, err := NewLocalDB(time.Millisecond)
				if err != nil {
					panic(err)
				}

				devAddr := lorawan.DevAddr([4]byte{1, 2, 3, 4})
				recipient := core.Recipient{Address: "MyAddress", Id: "MyId"}

				err = localDB.store(devAddr, recipient)
				So(err, ShouldBeNil)

				time.Sleep(10 * time.Millisecond)

				recipients, err := localDB.lookup(devAddr)
				So(recipients, ShouldBeNil)
				So(err, ShouldNotBeNil)

				recipients, err = localDB.lookup([4]byte{4, 5, 6, 7})
				So(err, ShouldNotBeNil)
				So(recipients, ShouldBeNil)
			})

			Convey("Store existing", func() {
				localDB, err := NewLocalDB(time.Hour)
				if err != nil {
					panic(err)
				}

				devAddr := lorawan.DevAddr([4]byte{1, 2, 3, 4})
				recipient := core.Recipient{Address: "MyAddress", Id: "MyId"}

				err = localDB.store(devAddr, recipient)
				So(err, ShouldBeNil)
				err = localDB.store(devAddr, recipient)
				So(err, ShouldNotBeNil)
			})
		})

	})
}
