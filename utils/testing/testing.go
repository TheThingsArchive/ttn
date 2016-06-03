// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package testing offers some handy methods to display check and cross symbols with colors in test
// logs.
package testing

import (
	"testing"

	"github.com/apex/log"
)

func GetLogger(t *testing.T, tag string) log.Interface {
	logger := &log.Logger{
		Handler: NewLogHandler(t),
		Level:   log.DebugLevel,
	}
	return logger.WithField("tag", tag)
}
