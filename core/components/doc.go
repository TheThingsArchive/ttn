// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package components offers implementations for all major components involved in the network.
//
// Router
//
// Routers are entry points of the network from the Nodes perspective. Packets transmitted by Nodes
// are forwarded to specific Routers from one or several Gateways. The Router then forwards those
// packets to one or several Brokers. The communication is bi-directional: Routers may also transfer
// packets from Broker to Gateways.
//
// Broker
//
// Brokers have a global vision of a network's part. They are in charge of several nodes, meaning
// that they will handle packets coming from those nodes (thereby, they are able to tell to Routers
// if they can handle a given packet). Several Routers may send packets coming from the same
// end-device (shared by several segments / Gateways), all duplicates are processed by the Broker
// and are sent to a corresponding Handler.
//
// A Broker is thereby able to check the integrity of a received packet and is closely communicating
// with a Network Server in order to administrate the related end-device. For a reference of
// magnitude, Brokers are designed to be in charge of a whole country or region (if the region has
// enough activity to deserve a dedicated Broker). Note that while brokers are able to verify the
// integrity of the packet (and therefore the identify of the end device), they are not able to read
// application data.
//
// Handler
//
// Handlers materialize the entry point to the network for Applications. They are secure referees
// which encode and decode data coming from application before transmitting them to a Broker of the
// network. Therefore, they are in charge of handling secret applications keys and only communicate
// an application id to Brokers as well as specific network session keys for each node (described in
// further sections). This way, the whole chain is able to forward a packet to the corresponding
// Handler without having any information about either the recipient (but a meaningless id) or the
// content.
//
// Because a given Handler is able to decrypt the data payload of a given packet, it could also
// implement mechanisms such as geolocation and send to the corresponding application some
// interesting meta-data alongside the data payload. Incidentally, a handler can only decrypt
// payload for packets related to applications registered to that handler. The handler is managing
// several secret application session keys and it uses these to encrypt and decrypt corresponding
// packet payloads.
//
// A Handler could be either part of an application or a standalone trusty server on which
// applications may register. The Things Network will provide Handlers as part of the whole network
// but - and this is true for any component - anyone could create its own implementation as long as
// it is compliant to the following specifications
//
// Network Controller
//
// Network controllers process MAC commands emitted by end-devices as well as taking care of the
// data rates and the frequency of the end-devices. They would emit commands to optimize
// the network by adjusting end-devices data rates / frequencies unless the node is requesting to
// keep its configuration as is.
//
// For the moment, a single Network controller will be associated for each Broker. No communication
// mechanisms between Network controllers is planned for the first version. Also, it won't be
// possible for a Broker to query another Network Server than the one it has been assigned to. Those
// features might be part of a second version.
package components
