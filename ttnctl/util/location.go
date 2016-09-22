// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"errors"
	"strconv"
	"strings"

	"github.com/TheThingsNetwork/ttn/core/account"
)

func ParseLocation(locationStr string) (*account.Location, error) {
	parts := strings.Split(locationStr, ",")
	if len(parts) != 2 {
		return nil, errors.New("Location should be on the <latitude>,<longitude> format")
	}

	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, err
	}

	if lat < 0 || lat > 90 {
		return nil, errors.New("Latitude should be in range [0, 90]")
	}

	lng, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, err
	}

	if lng < 0 || lng > 180 {
		return nil, errors.New("Longitude should be in range [0, 180]")
	}

	return &account.Location{
		Latitude:  lat,
		Longitude: lng,
	}, nil
}
