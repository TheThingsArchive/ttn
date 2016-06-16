package handler

import (
	"errors"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
)

// UplinkProcessor processes an uplink protobuf to an application-layer uplink message
type UplinkProcessor func(ctx log.Interface, ttnUp *pb_broker.DeduplicatedUplinkMessage, appUp *mqtt.UplinkMessage) error

// DownlinkProcessor processes an application-layer downlink message to a downlik protobuf
type DownlinkProcessor func(ctx log.Interface, appDown *mqtt.DownlinkMessage, ttnDown *pb_broker.DownlinkMessage) error

var ErrNotNeeded = errors.New("Further processing not needed")
