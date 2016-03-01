// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package readwriter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

type Interface interface {
	Write(data interface{})
	TryRead(to func(data []byte) error)
	Read(to func(data []byte))
	Bytes() ([]byte, error)
	Err() error
}

// entryReadWriter offers convenient method to write and read successively from a bytes buffer.
type rw struct {
	err  error
	data *bytes.Buffer
}

// newEntryReadWriter create a new read/writer from an existing buffer.
//
// If a nil or empty buffer is supplied, reading from the read/writer will cause an error (io.EOF)
// Nevertheless, if a valid non-empty buffer is given, the read/writer will start reading from the
// beginning of that buffer, and will start writting at the end of it.
func New(buf []byte) Interface {
	return &rw{
		err:  nil,
		data: bytes.NewBuffer(buf),
	}
}

// Write appends the given data at the end of the existing buffer.
//
// It does nothing if an error was previously noticed and panics if the given data are something
// different from: byte, []byte, string, AES128Key, EUI64, DevAddr.
//
// Also, it writes the length of the given raw data encoded on 2 bytes before writting the data
// itself. In that way, data can be appended and read easily.
func (w *rw) Write(data interface{}) {
	var dataLen uint16
	switch data.(type) {
	case uint8:
		dataLen = 1
	case uint16:
		dataLen = 2
	case uint32:
		dataLen = 4
	case uint64:
		dataLen = 8
	case []byte:
		dataLen = uint16(len(data.([]byte)))
	case lorawan.AES128Key:
		dataLen = uint16(len(data.(lorawan.AES128Key)))
	case lorawan.EUI64:
		dataLen = uint16(len(data.(lorawan.EUI64)))
	case lorawan.DevAddr:
		dataLen = uint16(len(data.(lorawan.DevAddr)))
	case string:
		str := data.(string)
		w.directWrite(uint16(len(str)))
		w.directWrite([]byte(str))
		return
	default:
		panic(fmt.Errorf("Unreckognized data type: %v", data))
	}
	w.directWrite(dataLen)
	w.directWrite(data)
}

// directWrite appends the given data at the end of the existing buffer (without the length).
func (w *rw) directWrite(data interface{}) {
	if w.err != nil {
		return
	}
	in := w.data.Next(w.data.Len())
	w.data = new(bytes.Buffer)
	binary.Write(w.data, binary.BigEndian, in)
	w.err = binary.Write(w.data, binary.BigEndian, data)
}

// Read retrieves next data from the given buffer. Implicitely, this implies the data to have been
// written using the Write method (len | data). Data are sent back through a callback as an array of
// bytes.
func (w *rw) Read(to func(data []byte)) {
	w.err = w.read(func(data []byte) error { to(data); return nil })
}

// TryRead retrieves the next data from the given buffer but differs from Read in the way it listen
// to the callback returns for a possible error.
func (w *rw) TryRead(to func(data []byte) error) {
	w.err = w.read(to)
}

func (w *rw) read(to func(data []byte) error) error {
	if w.err != nil {
		return w.err
	}

	lenTo := new(uint16)
	if err := binary.Read(w.data, binary.BigEndian, lenTo); err != nil {
		return err
	}
	next := w.data.Next(int(*lenTo))
	if len(next) != int(*lenTo) {
		return errors.New(errors.Structural, "Not enough data to read")
	}
	return to(next)
}

// Bytes might be used to retrieves the raw buffer after successive writes. It will return nil and
// an error if any issue was encountered during the process.
func (w rw) Bytes() ([]byte, error) {
	if w.err != nil {
		return nil, w.Err()
	}
	return w.data.Bytes(), nil
}

// Err just return the err status of the read-writer.
func (w rw) Err() error {
	if w.err != nil {
		if w.err == io.EOF {
			return errors.New(errors.Behavioural, w.err)
		}
		if failure, ok := w.err.(errors.Failure); ok {
			return failure
		}
		return errors.New(errors.Operational, w.err)
	}
	return nil
}
