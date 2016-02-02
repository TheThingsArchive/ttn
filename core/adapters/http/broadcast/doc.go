// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package broadcast extends the basic http adapter to allow registrations and broadcasting of a
// packet.
//
// Registrations are implicit. They are generated during broadcast depending on the response of all
// recipients. Technically, only one recipient is expected to give a positive answer (if we put
// aside malicious one). For that positive response would be generated a given registration.
//
// Recipients to whom broadcast packets are defined once during the adapter's instantiation.
package broadcast
