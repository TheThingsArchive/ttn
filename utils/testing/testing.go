// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package testing offers some handy methods to display check and cross symbols with colors in test
// logs.
package testing

import (
	"fmt"
	"testing"

	"github.com/apex/log"
)

func GetLogger(t *testing.T, tag string) log.Interface {
	logger := &log.Logger{
		Handler: NewLogHandler(t),
		Level:   log.DebugLevel,
	}
	return logger.WithField("tag", "Adapter")
}

// Ok displays a green check symbol
func Ok(t *testing.T, tag string) {
	t.Log(fmt.Sprintf("\033[32;1m\u2714 ok | %s\033[0m", tag))
}

// Ko fails the test and display a red cross symbol
func Ko(t *testing.T, format string, a ...interface{}) {
	t.Fatalf("\033[31;1m\u2718 ko | \033[0m\033[31m%s\033[0m", fmt.Sprintf(format, a...))
}

// Desc displays the provided description in cyan
func Desc(t *testing.T, format string, a ...interface{}) {
	t.Logf("\033[36m%s\033[0m", fmt.Sprintf(format, a...))
}
