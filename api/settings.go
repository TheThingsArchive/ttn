// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"time"
)

// WaitForStreams indicates how long should be waited for a stream to start
var WaitForStreams = 100 * time.Millisecond
