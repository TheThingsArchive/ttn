// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package dutycycle

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	errutil "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

var dutyManagerDB = path.Join(os.TempDir(), "TestDutyCycleStorage.db")

func TestGetSubBand(t *testing.T) {
	{
		Desc(t, "Test EuropeG")
		sb, err := GetSubBand(867.127)
		errutil.CheckErrors(t, nil, err)
		CheckSubBands(t, EuropeG, sb)
	}

	// --------------------

	{
		Desc(t, "Test EuropeG1")
		sb, err := GetSubBand(868.38)
		errutil.CheckErrors(t, nil, err)
		CheckSubBands(t, EuropeG1, sb)
	}

	// --------------------

	{
		Desc(t, "Test EuropeG2")
		sb, err := GetSubBand(868.865)
		errutil.CheckErrors(t, nil, err)
		CheckSubBands(t, EuropeG2, sb)
	}

	// --------------------

	{
		Desc(t, "Test EuropeG3")
		sb, err := GetSubBand(869.567)
		errutil.CheckErrors(t, nil, err)
		CheckSubBands(t, EuropeG3, sb)
	}

	// --------------------

	{
		Desc(t, "Test EuropeG4")
		sb, err := GetSubBand(869.856)
		errutil.CheckErrors(t, nil, err)
		CheckSubBands(t, EuropeG4, sb)
	}

	// --------------------

	{
		Desc(t, "Test Unknown")
		sb, err := GetSubBand(433.5)
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckSubBands(t, "", sb)
	}
}

func TestNewManager(t *testing.T) {
	defer func() {
		os.Remove(dutyManagerDB)
	}()

	{
		Desc(t, "Europe with valid cycleLength")
		m, err := NewManager(dutyManagerDB, time.Minute, Europe)
		errutil.CheckErrors(t, nil, err)
		err = m.Close()
		errutil.CheckErrors(t, nil, err)
	}

	// --------------------

	{
		Desc(t, "Europe with invalid cycleLength")
		_, err := NewManager(dutyManagerDB, 0, Europe)
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)
	}

	// --------------------

	{
		Desc(t, "Not europe with valid cycleLength")
		_, err := NewManager(dutyManagerDB, time.Minute, China)
		errutil.CheckErrors(t, pointer.String(string(errors.Implementation)), err)
	}
}

func TestUpdateAndLookup(t *testing.T) {
	defer func() {
		os.Remove(dutyManagerDB)
	}()

	{
		Desc(t, "Update unsupported frequency")

		// Build
		m, _ := NewManager(dutyManagerDB, time.Minute, Europe)

		// Operate
		err := m.Update([]byte{1, 2, 3}, 433.65, 100, "SF8BW125", "4/5")

		// Check
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)

		// Clean
		m.Close()
	}

	// --------------------

	{
		Desc(t, "Update invalid datr")

		// Build
		m, _ := NewManager(dutyManagerDB, time.Minute, Europe)

		// Operate
		err := m.Update([]byte{1, 2, 3}, 868.1, 100, "SF3BW125", "4/5")

		// Check
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)

		// Clean
		m.Close()
	}

	// --------------------

	{
		Desc(t, "Update invalid codr")

		// Build
		m, _ := NewManager(dutyManagerDB, time.Minute, Europe)

		// Operate
		err := m.Update([]byte{1, 2, 3}, 869.5, 100, "SF8BW125", "14")

		// Check
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)

		// Clean
		m.Close()
	}

	// --------------------

	{
		Desc(t, "Update once then lookup")

		// Build
		m, _ := NewManager(dutyManagerDB, time.Minute, Europe)

		// Operate
		err := m.Update([]byte{1, 2, 3}, 868.5, 14, "SF8BW125", "4/5")
		errutil.CheckErrors(t, nil, err)
		bands, err := m.Lookup([]byte{1, 2, 3})

		// Expectation
		want := map[subBand]uint{
			EuropeG1: 10,
		}

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckUsages(t, want, bands)

		// Clean
		m.Close()
	}

	// --------------------

	{
		Desc(t, "Update several then lookup")

		// Build
		m, _ := NewManager(dutyManagerDB, time.Minute, Europe)

		// Operate
		err := m.Update([]byte{4, 5, 6}, 868.523, 14, "SF8BW125", "4/5")
		errutil.CheckErrors(t, nil, err)
		err = m.Update([]byte{4, 5, 6}, 868.123, 42, "SF9BW125", "4/5")
		errutil.CheckErrors(t, nil, err)
		err = m.Update([]byte{4, 5, 6}, 867.785, 42, "SF8BW125", "4/6")
		errutil.CheckErrors(t, nil, err)
		bands, err := m.Lookup([]byte{4, 5, 6})

		// Expectation
		want := map[subBand]uint{
			EuropeG1: 51,
			EuropeG:  25,
		}

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckUsages(t, want, bands)

		// Clean
		m.Close()
	}

	// --------------------

	{
		Desc(t, "Update out of cycle then lookup")

		// Build
		m, _ := NewManager(dutyManagerDB, 250*time.Millisecond, Europe)

		// Operate
		err := m.Update([]byte{16, 2, 3}, 868.523, 14, "SF8BW125", "4/7")
		errutil.CheckErrors(t, nil, err)
		<-time.After(300 * time.Millisecond)
		err = m.Update([]byte{16, 2, 3}, 868.123, 42, "SF9BW125", "4/5")
		errutil.CheckErrors(t, nil, err)
		bands, err := m.Lookup([]byte{16, 2, 3})

		// Expectation
		want := map[subBand]uint{
			EuropeG1: 9871,
		}

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckUsages(t, want, bands)

		// Clean
		m.Close()
	}

	// --------------------

	{
		Desc(t, "Lookup out of cycle")

		// Build
		m, _ := NewManager(dutyManagerDB, time.Millisecond, Europe)

		// Operate
		err := m.Update([]byte{1, 2, 35}, 868.523, 14, "SF8BW125", "4/8")
		errutil.CheckErrors(t, nil, err)
		<-time.After(300 * time.Millisecond)
		bands, err := m.Lookup([]byte{1, 2, 35})

		// Expectation
		want := map[subBand]uint{}

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckUsages(t, want, bands)

		// Clean
		m.Close()
	}

	// -------------------

	{
		Desc(t, "Update on sf11 et sf12 with a 125 bandwidth -> optimization for low datarate")

		// Build
		m, _ := NewManager(dutyManagerDB, time.Minute, Europe)

		// Operate
		err := m.Update([]byte{4, 12, 6}, 868.523, 14, "SF11BW125", "4/5")
		errutil.CheckErrors(t, nil, err)
		err = m.Update([]byte{4, 12, 6}, 868.523, 42, "SF12BW125", "4/5")
		errutil.CheckErrors(t, nil, err)
		bands, err := m.Lookup([]byte{4, 12, 6})

		// Expectation
		want := map[subBand]uint{
			EuropeG1: 384,
		}

		// Check
		errutil.CheckErrors(t, nil, err)
		CheckUsages(t, want, bands)

		// Clean
		m.Close()
	}
}

func TestStateFromDuty(t *testing.T) {
	Desc(t, "Duty = 100 -> Blocked")
	CheckStates(t, StateBlocked, StateFromDuty(100))

	Desc(t, "Duty = 324 -> Blocked")
	CheckStates(t, StateBlocked, StateFromDuty(324))

	Desc(t, "Duty = 90 -> Warning")
	CheckStates(t, StateWarning, StateFromDuty(90))

	Desc(t, "Duty = 50 -> Available")
	CheckStates(t, StateAvailable, StateFromDuty(50))

	Desc(t, "Duty = 3 -> Highly Available")
	CheckStates(t, StateHighlyAvailable, StateFromDuty(3))
}
