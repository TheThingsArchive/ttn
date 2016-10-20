// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/apex/log"
)

// UplinkProcessor processes an uplink protobuf to an application-layer uplink message
type UplinkProcessor func(ctx log.Interface, ttnUp *pb_broker.DeduplicatedUplinkMessage, appUp *types.UplinkMessage) error

// DownlinkProcessor processes an application-layer downlink message to a downlik protobuf
type DownlinkProcessor func(ctx log.Interface, appDown *types.DownlinkMessage, ttnDown *pb_broker.DownlinkMessage) error

// ErrNotNeeded indicates that the processing of a message should be aborted
var ErrNotNeeded = errors.New("Further processing not needed")
