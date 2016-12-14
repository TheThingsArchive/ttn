// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package testing

import (
	"bytes"
	"sort"
	"sync"
	"testing"

	"fmt"

	"github.com/apex/log"
)

// colors.
const (
	none   = 0
	red    = 31
	green  = 32
	yellow = 33
	blue   = 34
	gray   = 90
)

// Colors mapping.
var Colors = [...]int{
	log.DebugLevel: gray,
	log.InfoLevel:  blue,
	log.WarnLevel:  yellow,
	log.ErrorLevel: red,
	log.FatalLevel: red,
}

// Strings mapping.
var Strings = [...]string{
	log.DebugLevel: "DEBUG",
	log.InfoLevel:  "INFO",
	log.WarnLevel:  "WARN",
	log.ErrorLevel: "ERROR",
	log.FatalLevel: "FATAL",
}

// field used for sorting.
type field struct {
	Name  string
	Value interface{}
}

// by sorts projects by call count.
type byName []field

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// LogHandler implementation.
type LogHandler struct {
	mu sync.Mutex
	T  *testing.T
}

// NewLogHandler handler.
func NewLogHandler(t *testing.T) *LogHandler {
	return &LogHandler{
		T: t,
	}
}

// HandleLog implements log.Handler.
func (h *LogHandler) HandleLog(e *log.Entry) error {
	color := Colors[e.Level]
	level := Strings[e.Level]

	var fields []field

	for k, v := range e.Fields {
		fields = append(fields, field{k, v})
	}

	sort.Sort(byName(fields))

	h.mu.Lock()
	defer h.mu.Unlock()

	buf := bytes.NewBuffer([]byte{})

	fmt.Fprintf(buf, "\033[%dm%6s\033[0m %-25s", color, level, e.Message)

	for _, f := range fields {
		fmt.Fprintf(buf, " \033[%dm%s\033[0m=%v", color, f.Name, f.Value)
	}

	h.T.Log(buf.String())

	return nil
}
