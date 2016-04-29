package storage

import (
	"encoding/hex"
	"errors"
	"strconv"
)

type HSlice struct {
	Data map[string]string
}

var (
	// ErrDoesNotExist indicates that a key does not exist
	ErrDoesNotExist = errors.New("HSlice: key does not exist")
	// ErrInvalidLength indicates that a slice had an invalid length
	ErrInvalidLength = errors.New("HSsice: invalid length")
)

func NewHSlice() *HSlice {
	return &HSlice{map[string]string{}}
}

// MarshalHSlice returns a slice with the data
func (s *HSlice) MarshalHSlice() (out []string) {
	for k, v := range s.Data {
		if k != "" && v != "" {
			out = append(out, k, v)
		}
	}
	return
}

// UnmarshalHSlice imports data from slice
func (s *HSlice) UnmarshalHSlice(slice []string) error {
	if len(slice)%2 != 0 {
		return ErrInvalidLength
	}
	for i := 0; i < len(slice); i = i + 2 {
		s.Data[slice[i]] = slice[i+1]
	}
	return nil
}

// SetFloat32 does what its name suggests
func (s *HSlice) SetFloat32(key string, value float32) {
	s.Data[key] = strconv.FormatFloat(float64(value), 'E', -1, 32)
}

// SetFloat64 does what its name suggests
func (s *HSlice) SetFloat64(key string, value float64) {
	s.Data[key] = strconv.FormatFloat(value, 'E', -1, 64)
}

// SetInt32 does what its name suggests
func (s *HSlice) SetInt32(key string, value int32) {
	s.SetInt64(key, int64(value))
}

// SetInt64 does what its name suggests
func (s *HSlice) SetInt64(key string, value int64) {
	s.Data[key] = strconv.FormatInt(value, 10)
}

// SetUint32 does what its name suggests
func (s *HSlice) SetUint32(key string, value uint32) {
	s.SetUint64(key, uint64(value))
}

// SetUint64 does what its name suggests
func (s *HSlice) SetUint64(key string, value uint64) {
	s.Data[key] = strconv.FormatUint(value, 10)
}

// SetBool does what its name suggests
func (s *HSlice) SetBool(key string, value bool) {
	s.Data[key] = strconv.FormatBool(value)
}

// SetString does what its name suggests
func (s *HSlice) SetString(key string, value string) {
	if value != "" {
		s.Data[key] = value
	}
}

// SetBytes does what its name suggests
func (s *HSlice) SetBytes(key string, value []byte) {
	if len(value) > 0 {
		s.Data[key] = hex.EncodeToString(value)
	}
}

// GetFloat32 does what its name suggests
func (s *HSlice) GetFloat32(key string) (value float32, err error) {
	if val, ok := s.Data[key]; ok {
		res, err := strconv.ParseFloat(val, 32)
		return float32(res), err
	}
	return 0, ErrDoesNotExist
}

// GetFloat64 does what its name suggests
func (s *HSlice) GetFloat64(key string) (value float64, err error) {
	if val, ok := s.Data[key]; ok {
		return strconv.ParseFloat(val, 64)
	}
	return 0, ErrDoesNotExist
}

// GetInt32 does what its name suggests
func (s *HSlice) GetInt32(key string) (value int32, err error) {
	if val, ok := s.Data[key]; ok {
		res, err := strconv.ParseInt(val, 10, 32)
		return int32(res), err
	}
	return 0, ErrDoesNotExist
}

// GetInt64 does what its name suggests
func (s *HSlice) GetInt64(key string) (value int64, err error) {
	if val, ok := s.Data[key]; ok {
		res, err := strconv.ParseInt(val, 10, 64)
		return res, err
	}
	return 0, ErrDoesNotExist
}

// GetUint32 does what its name suggests
func (s *HSlice) GetUint32(key string) (value uint32, err error) {
	if val, ok := s.Data[key]; ok {
		res, err := strconv.ParseUint(val, 10, 32)
		return uint32(res), err
	}
	return 0, ErrDoesNotExist
}

// GetUint64 does what its name suggests
func (s *HSlice) GetUint64(key string) (value uint64, err error) {
	if val, ok := s.Data[key]; ok {
		return strconv.ParseUint(val, 10, 64)
	}
	return 0, ErrDoesNotExist
}

// GetBool does what its name suggests
func (s *HSlice) GetBool(key string) (value bool, err error) {
	if val, ok := s.Data[key]; ok {
		return strconv.ParseBool(val)
	}
	return false, ErrDoesNotExist
}

// GetString does what its name suggests
func (s *HSlice) GetString(key string) (value string, err error) {
	if val, ok := s.Data[key]; ok {
		return val, nil
	}
	return "", ErrDoesNotExist
}

// GetBytes does what its name suggests
func (s *HSlice) GetBytes(key string) (value []byte, err error) {
	if val, ok := s.Data[key]; ok {
		return hex.DecodeString(val)
	}
	return []byte{}, ErrDoesNotExist
}
