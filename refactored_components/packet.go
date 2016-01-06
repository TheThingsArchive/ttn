// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/thethingsnetwork/core"
	//"github.com/thethingsnetwork/core/lorawan"
	"fmt"
	"github.com/thethingsnetwork/core/semtech"
)

var ErrImpossibleConversion = fmt.Errorf("The given packet can't be converted")

func ConvertRXPK(p semtech.RXPK) (core.Packet, error) {
	return core.Packet{}, nil
}
