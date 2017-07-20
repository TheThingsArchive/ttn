// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	pb_discovery "github.com/TheThingsNetwork/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Discover is used to discover another component
func (c *Component) Discover(serviceName, id string) (*pb_discovery.Announcement, error) {
	res, err := c.Discovery.Get(serviceName, id)
	if err != nil {
		return nil, errors.Wrapf(errors.FromGRPCError(err), "Failed to discover %s/%s", serviceName, id)
	}
	return res, nil
}

// Announce the component to TTN discovery
func (c *Component) Announce() error {
	if c.Identity.ID == "" {
		return errors.NewErrInvalidArgument("Component ID", "can not be empty")
	}
	err := c.Discovery.Announce(c.AccessToken)
	if err != nil {
		return errors.Wrapf(errors.FromGRPCError(err), "Failed to announce this component to TTN discovery: %s", err.Error())
	}
	c.Ctx.Info("ttn: Announced to TTN discovery")

	return nil
}
