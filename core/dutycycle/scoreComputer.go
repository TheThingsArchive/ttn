// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package dutycycle

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// ScoreComputer enables an external component to manipulate metadata associated to several targets
// in order to determine which target is the most suitable for a downlink response.
// It considers two windows RX1 and RX2 with the following conventions:
//
// For SF7 & SF8, RX1, the algorithm favors RX1
//
// For SF9+ or, if no target is available on RX1, then RX2 is used
//
// Within RX1 or RX2, the SNR is considered first (the higher the better), then the RSSI on a lower
// plan.
type ScoreComputer struct {
	sf uint
}

// BestTarget represents the best result that has been computed after all updates.
type BestTarget struct {
	ID    int  // The ID provided during updates
	IsRX2 bool // Whether it should use RX2
}

type scores struct {
	rx1 struct {
		ID    int
		Score int
	}
	rx2 struct {
		ID    int
		Score int
	}
}

// NewScoreComputer constructs a new ScoreComputer and initiate an empty scores table
func NewScoreComputer(datr string) (*ScoreComputer, scores, error) {
	sf, _, err := ParseDatr(datr)
	if err != nil {
		return nil, scores{}, errors.New(errors.Structural, err)
	}

	return &ScoreComputer{sf: uint(sf)}, scores{}, nil
}

// Update computes the score associated to the given target and update the internal score
// accordingly whether it is better than the existing one
func (c *ScoreComputer) Update(s scores, id int, metadata core.Metadata) scores {
	dutyRX1, dutyRX2 := metadata.DutyRX1, metadata.DutyRX2
	lsnr, rssi := float64(metadata.Lsnr), int(metadata.Rssi)

	rx1 := computeScore(State(dutyRX1), lsnr, rssi)
	if rx1 > s.rx1.Score {
		s.rx1.Score, s.rx1.ID = rx1, id
	}

	rx2 := computeScore(State(dutyRX2), lsnr, rssi)
	if rx2 > s.rx2.Score {
		s.rx2.Score, s.rx2.ID = rx2, id
	}

	return s
}

// Get returns the best score according to the configured spread factor and all updates.
// It returns nil if none of the target is available for a response
func (c *ScoreComputer) Get(s scores) *BestTarget {
	if s.rx1.Score > 0 && (c.sf == 7 || c.sf == 8) { // Favor RX1 on SF7 & SF8
		return &BestTarget{ID: s.rx1.ID, IsRX2: false}
	}
	if s.rx2.Score > 0 {
		return &BestTarget{ID: s.rx2.ID, IsRX2: true}
	}
	return nil
}

func computeScore(duty State, lsnr float64, rssi int) int {
	var score int

	// Most importance on the duty cycle
	switch duty {
	case StateHighlyAvailable:
		score += 5000
	case StateAvailable:
		score += 4900
	case StateWarning:
		score += 3000
	case StateBlocked:
		return 0
	}

	// Then, we favor lsnr
	if lsnr > 5.0 {
		score += 1000
	} else if lsnr > 4.0 {
		score += 750
	} else if lsnr > 3.0 {
		score += 500
	}

	// Rssi, negative will lower the score again.
	// For similar lsnr, this will make the difference.
	// We don't expect rssi to be lower than 500 though ...
	score += rssi

	return score
}
