// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package core contains the core library of The Things Network.
//
// This library can be used to built an entire network or only a small part of it. It is mainly
// divided in three parts:
//
// Packet manipulation
//
// The core package itself defines packets as we see them in the network as well as methods to
// serialize, convert and represent them.
//
// Because packets are likely to change over requests, this package centralizes all definitions and
// information related to them and share by the whole network. Each component may take and use only
// what it needs to operate over packets.
//
//
// Adapters
//
// The subfolder adapters hold all protocol adapters one could use to make components communicate.
// Each adapter implement its own protocol based on a given transport or application layer such as
// UDP, HTTP, TCP, CoAP or any fantasy needed.
//
//
// Components
//
// In the subfolder components you may find implementation of core logic of the network. The
// communication process has been abstracted by the adapters in such a way that all components only
// care about the business logic and the packet management.
//
// Components are split in 4 categories: router, broker, handler and network controller. Refer to
// the related documentation to find more information.
package core
