package types

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type DataRate struct {
	SpreadingFactor uint `json:"spreading_factor,omitempty"`
	Bandwidth       uint `json:"bandwidth,omitempty"`
}

// ParseDataRate parses a 32-bit hex-encoded string to a Devdatr
func ParseDataRate(input string) (datr *DataRate, err error) {
	re := regexp.MustCompile("SF(7|8|9|10|11|12)BW(125|250|500)")
	matches := re.FindStringSubmatch(input)
	if len(matches) != 3 {
		return nil, errors.New("ttn/core: Invalid DataRate")
	}

	sf, _ := strconv.ParseUint(matches[1], 10, 64)
	bw, _ := strconv.ParseUint(matches[2], 10, 64)

	return &DataRate{
		SpreadingFactor: uint(sf),
		Bandwidth:       uint(bw),
	}, nil
}

// Bytes returns the DataRate as a byte slice
func (datr DataRate) Bytes() []byte {
	return []byte(datr.String())
}

// String implements the Stringer interface.
func (datr DataRate) String() string {
	return fmt.Sprintf("SF%dBW%d", datr.SpreadingFactor, datr.Bandwidth)
}

// GoString implements the GoStringer interface.
func (datr DataRate) GoString() string {
	return datr.String()
}

// MarshalText implements the TextMarshaler interface.
func (datr DataRate) MarshalText() ([]byte, error) {
	return []byte(datr.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (datr *DataRate) UnmarshalText(data []byte) error {
	parsed, err := ParseDataRate(string(data))
	if err != nil {
		return err
	}
	*datr = *parsed
	return nil
}

// MarshalBinary implements the BinaryMarshaler interface.
func (datr DataRate) MarshalBinary() ([]byte, error) {
	return datr.MarshalText()
}

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (datr *DataRate) UnmarshalBinary(data []byte) error {
	return datr.UnmarshalText(data)
}

// Marshal implements the Marshaler interface.
func (datr DataRate) Marshal() ([]byte, error) {
	return datr.MarshalBinary()
}

// Unmarshal implements the Unmarshaler interface.
func (datr *DataRate) Unmarshal(data []byte) error {
	*datr = DataRate{} // Reset the receiver
	return datr.UnmarshalBinary(data)
}
