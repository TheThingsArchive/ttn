// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"sort"
	"strings"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/fields"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/toa"
)

type downlinkOption struct {
	gatewayPreference bool
	uplinkMetadata    *pb_gateway.RxMetadata
	option            *pb.DownlinkOption
}

// ByScore is used to sort a list of DownlinkOptions based on Score
type ByScore []downlinkOption

func (a ByScore) Len() int      { return len(a) }
func (a ByScore) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool {
	var pointsI, pointsJ int

	gradeBool := func(i, j bool, weight int) {
		if i {
			pointsI += weight
		}
		if j {
			pointsJ += weight
		}
	}

	gradeHighest := func(i, j float32, weight int) {
		if i > j {
			pointsI += weight
		}
		if i < j {
			pointsJ += weight
		}
	}

	gradeLowest := func(i, j float32, weight int) {
		if i < j {
			pointsI += weight
		}
		if i > j {
			pointsJ += weight
		}
	}

	// TODO: Score is deprecated, remove it
	if a[i].option.Score != 0 && a[j].option.Score != 0 {
		gradeLowest(float32(a[i].option.Score), float32(a[j].option.Score), 10)
		return pointsI > pointsJ
	}

	gradeBool(a[i].gatewayPreference, a[j].gatewayPreference, 10)
	gradeHighest(a[i].uplinkMetadata.Snr, a[j].uplinkMetadata.Snr, 1)
	gradeHighest(a[i].uplinkMetadata.Rssi, a[j].uplinkMetadata.Rssi, 1)
	gradeLowest(float32(a[i].option.PossibleConflicts), float32(a[j].option.PossibleConflicts), 1)
	gradeLowest(a[i].option.Utilization, a[j].option.Utilization, 1)
	gradeLowest(a[i].option.DutyCycle, a[j].option.DutyCycle, 1)

	toaI, _ := toa.Compute(a[i].option)
	toaJ, _ := toa.Compute(a[j].option)
	gradeLowest(float32(toaI.Seconds()), float32(toaJ.Seconds()), 1)

	return pointsI > pointsJ
}

func selectBestDownlink(options []downlinkOption) *pb.DownlinkOption {
	scored := ByScore(options)
	sort.Sort(scored)
	return scored[0].option
}

func (b *broker) HandleDownlink(downlink *pb.DownlinkMessage) (err error) {
	ctx := b.Ctx.WithFields(fields.Get(downlink))
	start := time.Now()
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle downlink")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled downlink")
		}
		if downlink != nil && b.monitorStream != nil {
			b.monitorStream.Send(downlink)
		}
	}()

	b.status.downlink.Mark(1)

	downlink.Trace = downlink.Trace.WithEvent(trace.ReceiveEvent)

	downlink, err = b.ns.Downlink(b.Component.GetContext(b.nsToken), downlink)
	if err != nil {
		return errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not handle downlink")
	}

	var routerID string
	if id := strings.Split(downlink.DownlinkOption.Identifier, ":"); len(id) == 2 {
		routerID = id[0]
	} else {
		return errors.NewErrInvalidArgument("DownlinkOption Identifier", "invalid format")
	}
	ctx = ctx.WithField("RouterID", routerID)

	router, err := b.getRouter(routerID)
	if err != nil {
		return err
	}

	downlink.Trace = downlink.Trace.WithEvent(trace.ForwardEvent, "router", routerID)

	router <- downlink

	return nil
}
