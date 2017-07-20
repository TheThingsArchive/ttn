// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"sync"
	"time"

	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	pb_router "github.com/TheThingsNetwork/api/router"
	"github.com/TheThingsNetwork/ttn/utils/toa"
	"github.com/rcrowley/go-metrics"
)

// Utilization manages the utilization of a gateway and its channels
// It is based on an exponentially weighted moving average over one minute
type Utilization interface {
	fmt.GoStringer
	// AddRx updates the utilization for receiving an uplink message
	AddRx(uplink *pb_router.UplinkMessage) error
	// AddRx updates the utilization for transmitting a downlink message
	AddTx(downlink *pb_router.DownlinkMessage) error
	// Get returns the overall rx and tx utilization for the gateway. If the gateway has multiple channels, the values will be 0 <= value < numChannels
	Get() (rx float64, tx float64)
	// GetChannel returns the rx and tx utilization for the given channel. The values will be 0 <= value < 1
	GetChannel(frequency uint64) (rx float64, tx float64)
	// Tick the clock to update the moving average. It should be called every 5 seconds
	Tick()
}

// NewUtilization creates a new Utilization
func NewUtilization() Utilization {
	return &utilization{
		overallRx: metrics.NewEWMA1(),
		channelRx: map[uint64]metrics.EWMA{},
		overallTx: metrics.NewEWMA1(),
		channelTx: map[uint64]metrics.EWMA{},
	}
}

type utilization struct {
	overallRx     metrics.EWMA
	channelRx     map[uint64]metrics.EWMA
	channelRxLock sync.RWMutex
	overallTx     metrics.EWMA
	channelTx     map[uint64]metrics.EWMA
	channelTxLock sync.RWMutex
}

func (u *utilization) GoString() (str string) {
	str += fmt.Sprintf("Rx %5.2f ", u.overallRx.Rate()/1000)
	for ch, r := range u.channelRx {
		str += fmt.Sprintf("(%d:%5.2f) ", ch, r.Rate()/1000)
	}
	str += "\n"
	str += fmt.Sprintf("Tx %5.2f ", u.overallTx.Rate()/1000)
	for ch, r := range u.channelTx {
		str += fmt.Sprintf("(%d:%5.2f) ", ch, r.Rate()/1000)
	}
	str += "\n"
	return
}

func (u *utilization) AddRx(uplink *pb_router.UplinkMessage) error {
	var t time.Duration
	var err error
	if lorawan := uplink.ProtocolMetadata.GetLoRaWAN(); lorawan != nil {
		if lorawan.Modulation == pb_lorawan.Modulation_LORA {
			t, err = toa.ComputeLoRa(uint(len(uplink.Payload)), lorawan.DataRate, lorawan.CodingRate)
			if err != nil {
				return err
			}
		}
		if lorawan.Modulation == pb_lorawan.Modulation_FSK {
			t, err = toa.ComputeFSK(uint(len(uplink.Payload)), int(lorawan.BitRate))
			if err != nil {
				return err
			}
		}
	}
	if t == 0 {
		return nil
	}
	u.overallRx.Update(int64(t) / 1000)
	frequency := uplink.GatewayMetadata.Frequency
	u.channelRxLock.Lock()
	defer u.channelRxLock.Unlock()
	if _, ok := u.channelRx[frequency]; !ok {
		u.channelRx[frequency] = metrics.NewEWMA1()
	}
	u.channelRx[frequency].Update(int64(t) / 1000)
	return nil
}

func (u *utilization) AddTx(downlink *pb_router.DownlinkMessage) error {
	var t time.Duration
	var err error
	if lorawan := downlink.ProtocolConfiguration.GetLoRaWAN(); lorawan != nil {
		if lorawan.Modulation == pb_lorawan.Modulation_LORA {
			t, err = toa.ComputeLoRa(uint(len(downlink.Payload)), lorawan.DataRate, lorawan.CodingRate)
			if err != nil {
				return err
			}
		}
		if lorawan.Modulation == pb_lorawan.Modulation_FSK {
			t, err = toa.ComputeFSK(uint(len(downlink.Payload)), int(lorawan.BitRate))
			if err != nil {
				return err
			}
		}
	}
	if t == 0 {
		return nil
	}
	u.overallTx.Update(int64(t) / 1000)
	frequency := downlink.GatewayConfiguration.Frequency
	u.channelTxLock.Lock()
	defer u.channelTxLock.Unlock()
	if _, ok := u.channelTx[frequency]; !ok {
		u.channelTx[frequency] = metrics.NewEWMA1()
	}
	u.channelTx[frequency].Update(int64(t) / 1000)
	return nil
}

func (u *utilization) Tick() {
	u.overallRx.Tick()
	u.channelRxLock.RLock()
	for _, ch := range u.channelRx {
		ch.Tick()
	}
	u.channelRxLock.RUnlock()
	u.overallTx.Tick()
	u.channelTxLock.RLock()
	for _, ch := range u.channelTx {
		ch.Tick()
	}
	u.channelTxLock.RUnlock()
}

func (u *utilization) Get() (float64, float64) {
	return u.overallRx.Snapshot().Rate() * 1000.0 / float64(time.Second), u.overallTx.Snapshot().Rate() * 1000.0 / float64(time.Second)
}

func (u *utilization) GetChannel(frequency uint64) (rx float64, tx float64) {
	u.channelRxLock.RLock()
	if channel, ok := u.channelRx[frequency]; ok {
		rx = channel.Snapshot().Rate() * 1000.0 / float64(time.Second)
	}
	u.channelRxLock.RUnlock()
	u.channelTxLock.RLock()
	if channel, ok := u.channelTx[frequency]; ok {
		tx = channel.Snapshot().Rate() * 1000.0 / float64(time.Second)
	}
	u.channelTxLock.RUnlock()
	return
}
