// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import "time"

// Proxies handle annoying json fields like datr and time
type payloadProxy struct {
	ProxRXPK []rxpkProxy `json:"rxpk,omitempty"`
	ProxStat *statProxy  `json:"stat,omitempty"`
	ProxTXPK *txpkProxy  `json:"txpk,omitempty"`
}

type statProxy struct {
	*Stat
	Time *timeparser `json:"time,omitempty"`
}

type rxpkProxy struct {
	*RXPK
	Datr *datrparser `json:"datr,omitempty"`
	Time *timeparser `json:"time,omitempty"`
}

type txpkProxy struct {
	*TXPK
	Datr *datrparser `json:"datr,omitempty"`
	Time *timeparser `json:"time,omitempty"`
}

// datrParser is used as a proxy to Unmarshal datr field in json payloads.
// Depending on the modulation type, the datr type could be either a string or a number.
// We're gonna parse it as a string in any case.
type datrparser struct {
	kind  string
	value *string // The parsed value
}

// timeParser is used as a proxy to Unmarshal JSON objects with different date types as the time
// module parse RFC3339 by default
type timeparser struct {
	layout string
	value  *time.Time // The parsed time value
}
