// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package http provides adapter implementations which run on top of http.
//
// The different protocols and mechanisms used are defined in the following document:
// https://github.com/TheThingsNetwork/ttn/blob/develop/documents/protocols/protocols.md
//
// The basic http adapter module can be used as a brick to build something bigger. By default, it
// does not hold registrations but only sending and reception of packets.
package http
