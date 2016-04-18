package core

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

func parseHEX(input string, length int) ([]byte, error) {
	pattern, err := regexp.Compile(fmt.Sprintf("[[:xdigit:]]{%d}", length))
	if err != nil {
		return nil, fmt.Errorf("Invalid pattern")
	}

	valid := pattern.MatchString(input)
	if !valid {
		return nil, fmt.Errorf("Invalid input: %s", input)
	}

	devAddr, err := hex.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("Could not decode input: %s", input)
	}

	return devAddr, nil
}

// ParseAddr parses a 32-bit hex-encoded string
func ParseAddr(input string) ([]byte, error) {
	return parseHEX(input, 8)
}

// ParseEUI parses a 64-bit hex-encoded string
func ParseEUI(input string) ([]byte, error) {
	return parseHEX(input, 16)
}

// ParseKey parses a 128-bit hex-encoded string
func ParseKey(input string) ([]byte, error) {
	return parseHEX(input, 32)
}

type AppEUI []byte
type DevEUI []byte
type DevAddr []byte

type AppKey []byte
type AppSKey []byte
type NwkSKey []byte

type DeviceType int

const (
	ABP DeviceType = iota
	OTAA
)

func (deviceType *DeviceType) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	switch {
	case str == "ABP":
		*deviceType = ABP
	case str == "OTAA":
		*deviceType = OTAA
	default:
		return fmt.Errorf("Unknown device type %s", str)
	}
	return nil
}

func (deviceType DeviceType) MarshalJSON() ([]byte, error) {
	switch deviceType {
	case OTAA:
		return []byte(`"OTAA"`), nil

	case ABP:
		return []byte(`"ABP"`), nil

	default:
		return nil, fmt.Errorf("Invalid device type value")
	}
}

func (appEUI AppEUI) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%X"`, appEUI)), nil
}

func (appEUI *AppEUI) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	parsed, err := ParseEUI(str)
	if err != nil {
		return err
	}

	*appEUI = AppEUI(parsed)
	return nil
}

func (devEUI DevEUI) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%X"`, devEUI)), nil
}

func (devEUI *DevEUI) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	parsed, err := ParseEUI(str)
	if err != nil {
		return err
	}

	*devEUI = DevEUI(parsed)
	return nil
}

func (devAddr DevAddr) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%X"`, devAddr)), nil
}

func (devAddr *DevAddr) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	parsed, err := ParseAddr(str)
	if err != nil {
		return err
	}

	*devAddr = DevAddr(parsed)
	return nil
}

func (appKey *AppKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%X"`, appKey)), nil
}

func (appKey *AppKey) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	parsed, err := ParseKey(str)
	if err != nil {
		return err
	}

	*appKey = AppKey(parsed)
	return nil
}

func (appSKey *AppSKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%X"`, appSKey)), nil
}

func (appSKey *AppSKey) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	parsed, err := ParseKey(str)
	if err != nil {
		return err
	}

	*appSKey = AppSKey(parsed)
	return nil
}

func (nwkSKey *NwkSKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%X"`, nwkSKey)), nil
}

func (nwkSKey *NwkSKey) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	parsed, err := ParseKey(str)
	if err != nil {
		return err
	}

	*nwkSKey = NwkSKey(parsed)
	return nil
}
