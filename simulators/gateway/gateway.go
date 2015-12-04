// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package gateway offers a dummy representation of a gateway.
//
// The package can be used to create a dummy gateway.
// Its former use is to provide a handy simulator for further testing of the whole network chain.
package gateway

import (
    "github.com/thethingsnetwork/core/lorawan/semtech"
)

type Gateway struct {
    Coord   GPSCoord                 // Gateway's GPS coordinates
    Routers []string                 // List of routers addresses

    ackr    float64                  // Percentage of upstream datagrams that were acknowledged
    dwnb    uint                     // Number of downlink datagrams received
    rxfw    uint                     // Number of radio packets forwarded
    rxnb    uint                     // Number of radio packets received
    rxok    uint                     // Number of radio packets received with a valid  PHY CRC
    txnb    uint                     // Number of packets emitted

    stderr  <-chan error             // Output error channel
    stdout  <-chan semtech.Packet    // Output communication channel
}

type GPSCoord struct {
    altitude    int     // GPS altitude in RX meters
    latitude    float64 // GPS latitude, North is +
    longitude   float64 // GPS longitude, East is +
}

func Create (id string, routers ...string) Gateway, error {
    return nil, nil
}

func genToken () []byte {
    return nil
}

type Forwarder interface {
    Forward(packet semtech.Packet) ()
    Mimic()
    Start() (<-chan semtech.Packet, <-chan error)
    Stat() semtech.Stat
}
