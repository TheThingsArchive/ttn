// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package readwriter

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/brocaar/lorawan"
)

// entryReadWriter offers convenient method to write and read successively from a bytes buffer.
type Interface struct {
	err  error
	data *bytes.Buffer
}

// newEntryReadWriter create a new read/writer from an existing buffer.
//
// If a nil or empty buffer is supplied, reading from the read/writer will cause an error (io.EOF)
// Nevertheless, if a valid non-empty buffer is given, the read/writer will start reading from the
// beginning of that buffer, and will start writting at the end of it.
func New(buf []byte) *Interface {
	return &Interface{
		err:  nil,
		data: bytes.NewBuffer(buf),
	}
}

// Write appends the given data at the end of the existing buffer.
//
// It does nothing if an error was previously noticed and panics if the given data are something
// different from: []byte, string, AES128Key, EUI64, DevAddr.
//
// Also, it writes the length of the given raw data encoded on 2 bytes before writting the data
// itself. In that way, data can be appended and read easily.
func (w *Interface) Write(data interface{}) {
	var raw []byte
	switch data.(type) {
	case []byte:
		raw = data.([]byte)
	case lorawan.AES128Key:
		data := data.(lorawan.AES128Key)
		raw = data[:]
	case lorawan.EUI64:
		data := data.(lorawan.EUI64)
		raw = data[:]
	case lorawan.DevAddr:
		data := data.(lorawan.DevAddr)
		raw = data[:]
	case string:
		raw = []byte(data.(string))
	default:
		panic(fmt.Errorf("Unreckognized data type: %v", data))
	}
	w.DirectWrite(uint16(len(raw)))
	w.DirectWrite(raw)
}

// DirectWrite appends the given data at the end of the existing buffer (without the length).
func (w *Interface) DirectWrite(data interface{}) {
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
func (w *Interface) Read(to func(data []byte)) {
	if w.err != nil {
		return
	}

	lenTo := new(uint16)
	if w.err = binary.Read(w.data, binary.BigEndian, lenTo); w.err != nil {
		return
	}
	to(w.data.Next(int(*lenTo)))
}

// Bytes might be used to retrieves the raw buffer after successive writes. It will return nil and
// an error if any issue was encountered during the process.
func (w Interface) Bytes() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}
	return w.data.Bytes(), nil
}

// Err just return the err status of the read-writer.
func (w Interface) Err() error {
	return w.err
}
