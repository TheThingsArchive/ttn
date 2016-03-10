// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package dutycycle

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/core/mocks"
)

// ----- CHECK utilities
func CheckSubBands(t *testing.T, want subBand, got subBand) {
	Check(t, want, got, "SubBands")
}

func CheckUsages(t *testing.T, want map[subBand]uint, got map[subBand]uint) {
	Check(t, want, got, "Usages")
}

func CheckBestTargets(t *testing.T, want *BestTarget, got *BestTarget) {
	Check(t, want, got, "Targets")
}

func CheckStates(t *testing.T, want State, got State) {
	Check(t, want, got, "States")
}
