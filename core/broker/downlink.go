// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"strings"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
)

// ByScore is used to sort a list of DownlinkOptions based on Score
type ByScore []*pb.DownlinkOption

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].Score < a[j].Score }

func (b *broker) HandleDownlink(downlink *pb.DownlinkMessage) error {
	ctx := b.Ctx.WithFields(log.Fields{
		"DevEUI": *downlink.DevEui,
		"AppEUI": *downlink.AppEui,
	})
	var err error
	start := time.Now()
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle downlink")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled downlink")
		}
	}()

	b.status.downlink.Mark(1)

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

	var router chan<- *pb.DownlinkMessage
	router, err = b.getRouter(routerID)
	if err != nil {
		return err
	}

	router <- downlink

	for _, monitor := range b.Monitors {
		ctx.Debug("Sending downlink to monitor")
		go monitor.SendDownlink(downlink)
	}

	return nil
}
