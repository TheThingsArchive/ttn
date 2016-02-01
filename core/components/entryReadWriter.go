// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"bytes"
	"encoding/binary"
)

type entryReadWriter struct {
	err  error
	data bytes.Buffer
}

func NewEntryReadWriter(buf []byte) *entryReadWriter {
	return &entryReadWriter{
		err:  nil,
		data: bytes.NewBuffer(buf),
	}
}

func (w *entryReadWriter) Write(data interface{}) {
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
