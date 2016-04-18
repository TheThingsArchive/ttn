package core

import (
	"encoding/hex"
	"fmt"
	"regexp"
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

func (appEUI AppEUI) String() string {
	return fmt.Println("%X", []byte(appEUI))
}

func (devEUI DevEUI) String() string {
	return fmt.Println("%X", []byte(devEUI))
}

func (devAddr DevAddr) String() string {
	return fmt.Println("%X", []byte(devAddr))
}

func (appKey AppKey) String() string {
	return fmt.Println("%X", []byte(appKey))
}

func (appSKey AppSKey) String() string {
	return fmt.Println("%X", []byte(appSKey))
}

func (nwkSKey NwkSKey) String() string {
	return fmt.Println("%X", []byte(nwkSKey))
}

func (deviceType *DeviceType) UnmarshalText(b []byte) error {
	str := string(b)
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

func (deviceType DeviceType) MarshalText() ([]byte, error) {
	switch deviceType {
	case OTAA:
		return []byte("OTAA"), nil

	case ABP:
		return []byte("ABP"), nil

	default:
		return nil, fmt.Errorf("Invalid device type value")
	}
}

func (appEUI AppEUI) MarshalText() ([]byte, error) {
	return []byte(appEUI.String()), nil
}

func (appEUI *AppEUI) UnmarshalText(data []byte) error {
	parsed, err := ParseEUI(string(data))
	if err != nil {
		return err
	}

	*appEUI = AppEUI(parsed)
	return nil
}

func (devEUI DevEUI) MarshalText() ([]byte, error) {
	return []byte(devEUI.String()), nil
}

func (devEUI *DevEUI) UnmarshalText(data []byte) error {
	parsed, err := ParseEUI(string(data))
	if err != nil {
		return err
	}

	*devEUI = DevEUI(parsed)
	return nil
}

func (devAddr DevAddr) MarshalText() ([]byte, error) {
	return []byte(devAddr.String()), nil
}

func (devAddr *DevAddr) UnmarshalText(data []byte) error {
	parsed, err := ParseAddr(string(data))
	if err != nil {
		return err
	}

	*devAddr = DevAddr(parsed)
	return nil
}

func (appKey *AppKey) MarshalText() ([]byte, error) {
	return []byte(appKey.String()), nil
}

func (appKey *AppKey) UnmarshalText(data []byte) error {
	parsed, err := ParseKey(string(data))
	if err != nil {
		return err
	}

	*appKey = AppKey(parsed)
	return nil
}

func (appSKey *AppSKey) MarshalText() ([]byte, error) {
	return []byte(appSKey.String()), nil
}

func (appSKey *AppSKey) UnmarshalText(data []byte) error {
	parsed, err := ParseKey(string(data))
	if err != nil {
		return err
	}

	*appSKey = AppSKey(parsed)
	return nil
}

func (nwkSKey *NwkSKey) MarshalText() ([]byte, error) {
	return []byte(nwkSKey.String()), nil
}

func (nwkSKey *NwkSKey) UnmarshalText(data []byte) error {
	parsed, err := ParseKey(string(data))
	if err != nil {
		return err
	}

	*nwkSKey = NwkSKey(parsed)
	return nil
}
