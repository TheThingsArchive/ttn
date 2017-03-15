// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"errors"
	"strconv"
	"strings"

	"github.com/TheThingsNetwork/go-account-lib/account"
)

func ParseLocation(locationStr string) (*account.AntennaLocation, error) {
	parts := strings.Split(locationStr, ",")
	if len(parts) != 2 {
		return nil, errors.New("Location should be on the <latitude>,<longitude> format")
	}

	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, err
	}

	if lat < -90 || lat > 90 {
		return nil, errors.New("Latitude should be in range [90, 90]")
	}

	lng, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, err
	}

	if lng < -180 || lng > 180 {
		return nil, errors.New("Longitude should be in range [-180, 180]")
	}

	return &account.AntennaLocation{
		Latitude:  &lat,
		Longitude: &lng,
	}, nil
}
