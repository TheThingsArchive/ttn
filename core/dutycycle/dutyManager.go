// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package dutycycle

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
	Update(id []byte, freq float64, size uint, datr string, codr string) error
	Lookup(id []byte) (Cycles, error)
	Close() error
}

type Cycles map[subBand]uint

type dutyManager struct {
	sync.RWMutex
	db           dbutil.Interface
	bucket       string
	CycleLength  time.Duration       // Duration upon which the duty-cycle is evaluated
	MaxDutyCycle map[subBand]float64 // The percentage max duty cycle accepted for each sub-band
}

// Available sub-bands
const (
	EuropeG  subBand = "europe g"
	EuropeG1         = "europe g1"
	EuropeG2         = "europe g2"
	EuropeG3         = "europe g3"
	EuropeG4         = "europe g4"
)

type subBand string

type State uint

const (
	StateHighlyAvailable State = iota
	StateAvailable
	StateWarning
	StateBlocked
)

// Available regions for LoRaWAN
const (
	Europe region = iota
	US
	China
)

type region byte

// GetSubBand returns the subband associated to a given frequency
func GetSubBand(freq float64) (subBand, error) {
	// g 865.0 – 868.0 MHz 1% or LBT+AFA, 25 mW (=14dBm)
	if freq >= 865.0 && freq < 868.0 {
		return EuropeG, nil
	}

	// g1 868.0 – 868.6 MHz 1% or LBT+AFA, 25 mW
	if freq >= 868.0 && freq < 868.6 {
		return EuropeG1, nil
	}

	// g2 868.7 – 869.2 MHz 0.1% or LBT+AFA, 25 mW
	if freq >= 868.7 && freq < 869.2 {
		return EuropeG2, nil
	}

	// g3 869.4 – 869.65 MHz 10% or LBT+AFA, 500 mW (=27dBm)
	if freq >= 869.4 && freq < 869.65 {
		return EuropeG3, nil
	}

	// g4 869.7 – 870.0 MHz 1% or LBT+AFA, 25 mW
	if freq >= 869.7 && freq < 870 {
		return EuropeG4, nil
	}

	return "", errors.New(errors.Structural, "Unknown frequency")
}

// NewManager constructs a new gateway manager from
func NewManager(filepath string, cycleLength time.Duration, r region) (DutyManager, error) {
	var maxDuty map[subBand]float64
	switch r {
	case Europe:
		maxDuty = map[subBand]float64{
			EuropeG:  0.01,
			EuropeG1: 0.01,
			EuropeG2: 0.001,
			EuropeG3: 0.1,
			EuropeG4: 0.01,
		}
	default:
		return nil, errors.New(errors.Implementation, "Region not supported")
	}

	if cycleLength == 0 {
		return nil, errors.New(errors.Structural, "Invalid cycleLength. Should be > 0")
	}

	// Try to start a database
	db, err := dbutil.New(filepath)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return &dutyManager{
		db:           db,
		bucket:       "cycles",
		CycleLength:  cycleLength,
		MaxDutyCycle: maxDuty,
	}, nil
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

	// Compute the ime-on-air
	timeOnAir, err := computeTOA(size, datr, codr)
	if err != nil {
		return err
	}

	// Lookup and update the entry
	m.Lock()
	defer m.Unlock()
	itf, err := m.db.Lookup(m.bucket, id, &dutyEntry{})

	var entry dutyEntry
	if err == nil {
		entry = itf.([]dutyEntry)[0]
	} else if err.(errors.Failure).Nature == errors.NotFound {
		entry = dutyEntry{
			Until: time.Unix(0, 0),
			OnAir: make(map[subBand]time.Duration),
		}
	} else {
		return err
	}

	// If the previous cycle is done, we create a new one
	if entry.Until.Before(time.Now()) {
		entry.Until = time.Now().Add(m.CycleLength)
		entry.OnAir[sub] = timeOnAir
	} else {
		entry.OnAir[sub] += timeOnAir
	}

	return m.db.Replace(m.bucket, id, []dbutil.Entry{&entry})
}

// Lookup returns the current bandwidth usages for a set of subband
//
// The usage is an integer between 0 and 100 (maybe above 100 if the usage exceed the limitation).
// The closest to 0, the more usage we have
func (m *dutyManager) Lookup(id []byte) (Cycles, error) {
	m.RLock()
	defer m.RUnlock()

	// Lookup the entry
	itf, err := m.db.Lookup(m.bucket, id, &dutyEntry{})
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

// Close releases the database access
func (m *dutyManager) Close() error {
	return m.db.Close()
}

// computeTOA computes the time-on-air given a size in byte, a LoRaWAN datr identifier, an LoRa Codr
// identifier.
func computeTOA(size uint, datr string, codr string) (time.Duration, error) {
	// Ensure the datr and codr are correct
	var rc float64
	switch codr {
	case "4/5":
		rc = 4.0 / 5.0
	case "4/6":
		rc = 4.0 / 6.0
	case "4/7":
		rc = 4.0 / 7.0
	case "4/8":
		rc = 4.0 / 8.0
	default:
		return 0, errors.New(errors.Structural, "Invalid Codr")
	}

	sf, bw, err := ParseDatr(datr)
	if err != nil {
		return 0, err
	}

	// Additional variables needed to compute times on air
	s := float64(size)
	var de float64
	if bw == 125 && (sf == 11 || sf == 12) {
		de = 1.0
	}

	// Compute toa, Page 7: http://www.semtech.com/images/datasheet/LoraDesignGuide_STD.pdf
	payloadNb := 8.0 + math.Max(0, 4*math.Ceil((2*s-sf-6)/(sf-2*de))/rc)
	timeOnAir := (payloadNb + 12.25) * math.Pow(2, sf) / bw // in ms

	return time.ParseDuration(fmt.Sprintf("%fms", timeOnAir))
}

func ParseDatr(datr string) (float64, float64, error) {
	re := regexp.MustCompile("^SF(7|8|9|10|11|12)BW(125|250|500)$")
	matches := re.FindStringSubmatch(datr)

	if len(matches) != 3 {
		return 0, 0, errors.New(errors.Structural, "Invalid Datr")
	}

	sf, _ := strconv.ParseFloat(matches[1], 64)
	bw, _ := strconv.ParseFloat(matches[2], 64)

	return sf, bw, nil
}

func StateFromDuty(duty uint) State {
	if duty >= 100 {
		return StateBlocked
	}

	if duty > 85 {
		return StateWarning
	}

	if duty > 30 {
		return StateAvailable
	}

	return StateHighlyAvailable

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
