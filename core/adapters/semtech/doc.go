// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package semtech provides an Adapter which implements the semtech forwarder protocol.
//
// The protocol could be found in this document:
// https://github.com/TheThingsNetwork/ttn/blob/develop/documents/protocols/semtech.pdf
//
// This protocol and to some extend the adapter does not allow a spontaneous downlink transmission
// to be initiated. Response packets are only transmitted if are a response to uplink packets. The
// AckNacker handles that logic.
package semtech
