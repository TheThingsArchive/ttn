// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/dutycycle"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

type scoreComputer struct {
	chrx1   chan inScore
	chrx2   chan inScore
	chflush chan chan *outScore
}

type inScore struct {
	ID    int
	Score int
}

type outScore struct {
	ID    int
	IsRX2 bool
}

func newScoreComputer(datr *string) (*scoreComputer, error) {
	if datr == nil {
		return nil, errors.New(errors.Structural, "Missing mandatory metadata datr")
	}
	sf, _, err := ParseDatr(*datr)
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	c := &scoreComputer{
		chrx1:   make(chan inScore),
		chrx2:   make(chan inScore),
		chflush: make(chan chan *outScore),
	}

	go watchUpdate(uint(sf), c.chrx1, c.chrx2, c.chflush)
	return c, nil
}

func (c *scoreComputer) Update(i int, metadata core.Metadata) {
	// And try to find the best recipient to which answer
	dutyRX1 := metadata.DutyRX1
	dutyRX2 := metadata.DutyRX2
	lsnr := metadata.Lsnr
	rssi := metadata.Rssi

	if dutyRX1 == nil || dutyRX2 == nil || lsnr == nil || rssi == nil {
		// NOTE We could possibly compute something if we had some of them (but not all).
		// However, for now, we expect them all
		return
	}

	c.chrx1 <- inScore{ID: i, Score: computeScore(State(*dutyRX1), *lsnr, *rssi)}
	c.chrx2 <- inScore{ID: i, Score: computeScore(State(*dutyRX2), *lsnr, *rssi)}
}

func (c *scoreComputer) Get() *outScore {
	chout := make(chan *outScore)
	c.chflush <- chout
	return <-chout
}

func watchUpdate(sf uint, chrx1 <-chan inScore, chrx2 <-chan inScore, chflush <-chan chan *outScore) {
	var sRX1, sRX2 inScore

	for {
		select {
		case rx1 := <-chrx1:
			if rx1.Score > sRX1.Score {
				sRX1 = rx1
			}
		case rx2 := <-chrx2:
			if rx2.Score > sRX2.Score {
				sRX2 = rx2
			}
		case flush := <-chflush:
			if sRX1.Score > 0 && (sf == 7 || sf == 8) { // Favor RX1 on SF7 & SF8
				flush <- &outScore{ID: sRX1.ID, IsRX2: false}
				return
			}
			if sRX2.Score > 0 {
				flush <- &outScore{ID: sRX2.ID, IsRX2: true}
				return
			}
			flush <- nil
			return
		}
	}
}

func computeScore(duty State, lsnr float64, rssi int) int {
	var score int

	switch duty {
	case StateHighlyAvailable:
		score += 3000
	case StateAvailable:
		score += 2000
	case StateWarning:
		score += 1000
	case StateBlocked:
		return 0
	}

	if lsnr > 5.0 {
		score += 200
	} else if lsnr > 3.0 {
		score += 100
	}

	score -= rssi

	return score
}
