// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package trace

// Event types
const (
	AcceptEvent        = "accept"
	BuildDownlinkEvent = "build downlink"
	CheckMICEvent      = "check mic"
	DeduplicateEvent   = "deduplicate"
	DropEvent          = "drop"
	ForwardEvent       = "forward"
	HandleMACEvent     = "handle mac command"
	ReceiveEvent       = "receive"
	SendEvent          = "send"
	UpdateStateEvent   = "update state"
)
