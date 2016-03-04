// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
)

// DutyManager provides an interface to manipulate and compute gateways duty-cycles.
type DutyManager interface {
}

type dutyManager struct {
	sync.RWMutex
	db           dbutil.Interface
	CycleLength  time.Duration       // Duration upon which the duty-cycle is evaluated
	MaxDutyCycle map[subBand]float64 // The percentage max duty cycle accepted for each sub-band
}

// Available sub-bands
const (
	EuropeRX1_A subBand = iota
	EuropeRX1_B
	EuropeRX2
)

type subBand byte

// Available regions for LoRaWAN
const (
	Europe region = iota
	US
	China
)

type region byte

// NewDutyManager constructs a new gateway manager from
func NewDutyManager(filepath string, cycleLength time.Duration, r region) (DutyManager, error) {
	var maxDuty map[subBand]float64
	switch r {
	case Europe:
		maxDuty = map[subBand]float64{
			EuropeRX1_A: 0.01, // 1% dutycycle
			EuropeRX1_B: 0.01, // 1% dutycycle
			EuropeRX2:   0.1,  // 10% dutycycle
		}
	default:
		return nil, errors.New(errors.Implementation, "Region not supported")
	}

	// Try to start a database
	db, err := dbutil.New(filepath)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return &dutyManager{
		db:           db,
		CycleLength:  cycleLength,
		MaxDutyCycle: maxDuty,
	}, nil
}

// GetSubBand returns the subband associated to a given frequency
func GetSubBand(freq float64) (subBand, error) {
	// EuropeRX1_A -> 868.1 MHz -> 868.9 MHz
	if int(freq) == 868 {
		return EuropeRX1_A, nil
	}

	// EuropeRX1_B -> 867.1 MHz -> 867.9 MHz
	if int(freq) == 869 {
		return EuropeRX1_B, nil
	}

	// EuropeRX2 -> 869.5 MHz
	if math.Floor(freq*10.0) == 8695.0 {
		return EuropeRX2, nil
	}
	return 0, errors.New(errors.Structural, "Unknown frequency")
}

// Update update an entry with the corresponding time-on-air
//
// Datr represents a LoRaWAN data-rate indicator of the form SFxxBWyyy,
// where xx C [[7;12]] and yyy C { 125, 250, 500 }
// Codr represents a LoRaWAN code rate  indicator fo the form 4/x with x C [[5;8]]
func (m *dutyManager) Update(id []byte, freq float64, size uint, datr string, codr string) error {
	sub, err := GetSubBand(freq)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%+x", id)

	// Compute the ime-on-air
	timeOnAir, err := computeTOA(size, datr, codr)
	if err != nil {
		return err
	}

	// Lookup and update the entry
	m.Lock()
	defer m.Unlock()
	itf, err := m.db.Lookup(key, []byte("entry"), &dutyEntry{})
	if err != nil {
		return err
	}
	entry := itf.([]dutyEntry)[0]

	// If the previous cycle is done, we create a new one
	if entry.Until.Before(time.Now()) {
		entry.Until = time.Now()
		entry.OnAir[sub] = timeOnAir
	} else {
		entry.OnAir[sub] += timeOnAir
	}

	return m.db.Replace(key, []byte("entry"), []dbutil.Entry{&entry})
}

// Lookup returns the current bandwidth usages for a set of subband
//
// The usage is an integer between 0 and 100 (maybe above 100 if the usage exceed the limitation).
// The closest to 0, the more usage we have
func (m *dutyManager) Lookup(id []byte) (map[subBand]uint, error) {
	m.RLock()
	defer m.RUnlock()

	// Lookup the entry
	itf, err := m.db.Lookup(fmt.Sprintf("%+x", id), []byte("entry"), &dutyEntry{})
	if err != nil {
		return nil, err
	}
	entry := itf.([]dutyEntry)[0]

	// For each sub-band, compute the remaining time-on-air available
	cycles := make(map[subBand]uint)
	if entry.Until.After(time.Now()) {
		for s, toa := range entry.OnAir {
			// The actual duty cycle
			dutyCycle := float64(toa.Nanoseconds()) / float64(m.CycleLength.Nanoseconds())
			// Now, how full are we comparing to the limitation, in percent
			cycles[s] = uint(100 * dutyCycle / m.MaxDutyCycle[s])
		}
	}

	return cycles, nil
}

// computeTOA computes the time-on-air given a size in byte, a LoRaWAN datr identifier, an LoRa Codr
// identifier.
func computeTOA(size uint, datr string, codr string) (time.Duration, error) {
	// Ensure the datr and codr are correct
	var cr float64
	switch codr {
	case "4/5":
		cr = 4.0 / 5.0
	case "4/6":
		cr = 4.0 / 6.0
	case "4/7":
		cr = 4.0 / 7.0
	case "4/8":
		cr = 4.0 / 8.0
	default:
		return 0, errors.New(errors.Structural, "Invalid Codr")
	}

	re := regexp.MustCompile("^SF(7|8|9|10|11|12)BW(125|250|500)$")
	matches := re.FindStringSubmatch(datr)

	if len(matches) != 3 {
		return 0, errors.New(errors.Structural, "Invalid Datr")
	}

	// Compute bitrate, Page 10: http://www.semtech.com/images/datasheet/an1200.22.pdf
	sf, _ := strconv.ParseFloat(matches[1], 64)
	bw, _ := strconv.ParseUint(matches[2], 10, 64)
	bitrate := sf * cr * float64(bw) / math.Pow(2, sf)

	return time.Duration(float64(size*8) / bitrate), nil
}

type dutyEntry struct {
	Until time.Time                 `json:"until"`
	OnAir map[subBand]time.Duration `json:"on_air"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e dutyEntry) MarshalBinary() ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	return data, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *dutyEntry) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, e); err != nil {
		return errors.New(errors.Structural, err)
	}
	return nil
}
