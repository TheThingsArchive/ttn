// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package dutycycle

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	errutil "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestNewScoreComputer(t *testing.T) {
	{
		Desc(t, "Nil datr as argument")
		_, _, err := NewScoreComputer(nil)
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)
	}

	// --------------------

	{
		Desc(t, "Invalid datr as argument")
		_, _, err := NewScoreComputer(pointer.String("TheThingsNetwork"))
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)
	}

	// --------------------

	{
		Desc(t, "Valid datr")
		_, _, err := NewScoreComputer(pointer.String("SF8BW250"))
		errutil.CheckErrors(t, nil, err)
	}
}

func TestUpdateGet(t *testing.T) {
	t.Log(`
	The following set of tests follows the given notation:
		Desc := SF | Upgrade
		SF := { SF7, SF8, SF9, SF10, SF11, SF12 }
		Upgrade := ... | UpgradeDesc
		UpgradeDesc := (ID, State, State, Rssi, Lsnr)
		ID := uint
		State := { Ha, Av, Wa, Bl }
		Rssi := int
		Lsnr := int
	`)

	{
		Desc(t, "SF7 | ...")

		// Build
		c, s, err := NewScoreComputer(pointer.String("SF7BW125"))
		errutil.CheckErrors(t, nil, err)

		// Operate
		got := c.Get(s)

		// Check
		CheckBestTargets(t, nil, got)
	}

	// --------------------

	{
		Desc(t, "SF7 | (1, Av, Bl, -25, 5.0)")

		// Build
		c, s, err := NewScoreComputer(pointer.String("SF7BW125"))
		errutil.CheckErrors(t, nil, err)

		// Operate
		s = c.Update(s, 1, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateAvailable)),
			DutyRX2: pointer.Uint(uint(StateBlocked)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(5.0),
		})
		got := c.Get(s)

		// Check
		CheckBestTargets(t, &BestTarget{ID: 1, IsRX2: false}, got)
	}

	// --------------------

	{
		Desc(t, "SF7 | (1, Bl, Ha, -25, 5.0)")

		// Build
		c, s, err := NewScoreComputer(pointer.String("SF7BW125"))
		errutil.CheckErrors(t, nil, err)

		// Operate
		s = c.Update(s, 1, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateBlocked)),
			DutyRX2: pointer.Uint(uint(StateHighlyAvailable)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(5.0),
		})
		got := c.Get(s)

		// Check
		CheckBestTargets(t, &BestTarget{ID: 1, IsRX2: true}, got)
	}

	// --------------------

	{
		Desc(t, "SF7 | (1, Bl, Bl, -25, 5.0)")

		// Build
		c, s, err := NewScoreComputer(pointer.String("SF7BW125"))
		errutil.CheckErrors(t, nil, err)

		// Operate
		s = c.Update(s, 1, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateBlocked)),
			DutyRX2: pointer.Uint(uint(StateBlocked)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(5.0),
		})
		got := c.Get(s)

		// Check
		CheckBestTargets(t, nil, got)
	}

	// --------------------

	{
		Desc(t, "SF9 | (1, Av, Av, -25, 5.0) ")

		// Build
		c, s, err := NewScoreComputer(pointer.String("SF9BW125"))
		errutil.CheckErrors(t, nil, err)

		// Operate
		s = c.Update(s, 1, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateAvailable)),
			DutyRX2: pointer.Uint(uint(StateAvailable)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(5.0),
		})
		got := c.Get(s)

		// Check
		CheckBestTargets(t, &BestTarget{ID: 1, IsRX2: true}, got)
	}

	// --------------------

	{
		Desc(t, "SF10 | (1, Av, Av, -25, 5.0) :: (2, Av, Av, -25, 3.0)")

		// Build
		c, s, err := NewScoreComputer(pointer.String("SF10BW125"))
		errutil.CheckErrors(t, nil, err)

		// Operate
		s = c.Update(s, 1, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateAvailable)),
			DutyRX2: pointer.Uint(uint(StateAvailable)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(5.0),
		})
		s = c.Update(s, 2, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateAvailable)),
			DutyRX2: pointer.Uint(uint(StateAvailable)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(3.0),
		})
		got := c.Get(s)

		// Check
		CheckBestTargets(t, &BestTarget{ID: 1, IsRX2: true}, got)
	}

	// --------------------

	{
		Desc(t, "SF10 | (1, Av, Bl, -25, 5.0)")

		// Build
		c, s, err := NewScoreComputer(pointer.String("SF10BW125"))
		errutil.CheckErrors(t, nil, err)

		// Operate
		s = c.Update(s, 1, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateAvailable)),
			DutyRX2: pointer.Uint(uint(StateBlocked)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(5.0),
		})
		got := c.Get(s)

		// Check
		CheckBestTargets(t, nil, got)
	}

	// --------------------

	{
		Desc(t, "SF8 | (1, Wa, Av, -25, 5.0) :: (2, Av, Av, -25, 5.0)")

		// Build
		c, s, err := NewScoreComputer(pointer.String("SF8BW125"))
		errutil.CheckErrors(t, nil, err)

		// Operate
		s = c.Update(s, 1, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateWarning)),
			DutyRX2: pointer.Uint(uint(StateAvailable)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(5.0),
		})
		s = c.Update(s, 2, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateAvailable)),
			DutyRX2: pointer.Uint(uint(StateAvailable)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(5.0),
		})
		got := c.Get(s)

		// Check
		CheckBestTargets(t, &BestTarget{ID: 2, IsRX2: false}, got)
	}

	// --------------------

	{
		Desc(t, "SF12 | (1, Av, Av, -25, 5.1) :: (2, Ha, Ha, -25, 3.4)")

		// Build
		c, s, err := NewScoreComputer(pointer.String("SF12BW125"))
		errutil.CheckErrors(t, nil, err)

		// Operate
		s = c.Update(s, 1, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateAvailable)),
			DutyRX2: pointer.Uint(uint(StateAvailable)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(5.1),
		})
		s = c.Update(s, 2, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateHighlyAvailable)),
			DutyRX2: pointer.Uint(uint(StateHighlyAvailable)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(3.4),
		})
		got := c.Get(s)

		// Check
		CheckBestTargets(t, &BestTarget{ID: 1, IsRX2: true}, got)
	}

	// --------------------

	{
		Desc(t, "Missing metadata in update")

		// Build
		c, s, err := NewScoreComputer(pointer.String("SF12BW125"))
		errutil.CheckErrors(t, nil, err)

		// Operate
		s = c.Update(s, 1, core.Metadata{
			DutyRX1: pointer.Uint(uint(StateAvailable)),
			Rssi:    pointer.Int(-25),
			Lsnr:    pointer.Float64(5.1),
		})
		got := c.Get(s)

		// Check
		CheckBestTargets(t, nil, got)
	}
}
