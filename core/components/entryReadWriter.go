// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/brocaar/lorawan"
)

type entryReadWriter struct {
	err  error
	data *bytes.Buffer
}

func NewEntryReadWriter(buf []byte) *entryReadWriter {
	return &entryReadWriter{
		err:  nil,
		data: bytes.NewBuffer(buf),
	}
}

func (w *entryReadWriter) Write(data interface{}) {
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

func (w *entryReadWriter) DirectWrite(data interface{}) {
	if w.err != nil {
		return
	}
	w.err = binary.Write(w.data, binary.BigEndian, data)
}

func (w *entryReadWriter) Read(to func(data []byte)) {
	if w.err != nil {
		return
	}
	lenTo := new(uint16)
	if w.err = binary.Read(w.data, binary.BigEndian, lenTo); w.err != nil {
		return
	}
	to(w.data.Next(int(*lenTo)))
}

func (w entryReadWriter) Bytes() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}
	return w.data.Bytes(), nil
}

func (w entryReadWriter) Err() error {
	return w.err
}
