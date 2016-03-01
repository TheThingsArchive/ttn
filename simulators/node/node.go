// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package node

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

func randBytes(n int) []byte {
	bytes := make([]byte, n)
	for i := range bytes {
		bytes[i] = byte(rand.Intn(255))
	}
	return bytes
}

type LiveNode interface {
	Start(chan string)
}

type Node struct {
	DevAddr  lorawan.DevAddr
	AppEUI   lorawan.EUI64
	NwkSKey  lorawan.AES128Key
	AppSKey  lorawan.AES128Key
	FCntUp   uint32
	interval int
	ctx      log.Interface
}

func New(interval int, ctx log.Interface) *Node {
	devAddr := [4]byte{}
	copy(devAddr[:], randBytes(4))

	appEUI := [8]byte{}
	copy(appEUI[:], randBytes(8))

	nwkSKey := [16]byte{}
	copy(nwkSKey[:], randBytes(16))

	appSKey := [16]byte{}
	copy(appSKey[:], randBytes(16))

	node := &Node{
		DevAddr:  lorawan.DevAddr(devAddr),
		AppEUI:   lorawan.EUI64(appEUI),
		NwkSKey:  lorawan.AES128Key(nwkSKey),
		AppSKey:  lorawan.AES128Key(appSKey),
		interval: interval,
	}

	node.ctx = ctx.WithField("devAddr", node.DevAddr)

	return node
}

func (node *Node) Start(messages chan string) {
	for {
		<-time.After(time.Duration(rand.ExpFloat64()*float64(node.interval)) * time.Millisecond)
		node.NextMessage(messages)
	}
}

func (node *Node) String() string {
	return fmt.Sprintf("Node %s:\n  AppEUI: %s\n  NwkSKey: %s\n  AppSKey: %s\n  FCntUp: %d", node.DevAddr, node.AppEUI, node.NwkSKey, node.AppSKey, node.FCntUp)
}

func (node *Node) NextMessage(messages chan string) {
	node.FCntUp++
	raw := node.BuildMessage([]byte(fmt.Sprintf("This is message %d.", node.FCntUp)))
	node.ctx.WithField("FCnt", node.FCntUp).Debug("Publishing message")
	messages <- strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")
}

func (node *Node) BuildMessage(data []byte) []byte {
	uplink := true

	macPayload := lorawan.NewMACPayload(uplink)
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: node.DevAddr,
		FCtrl: lorawan.FCtrl{
			ADR:       false,
			ADRACKReq: false,
			ACK:       false,
		},
		FCnt: node.FCntUp,
	}
	macPayload.FPort = 10
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: data}}

	if err := macPayload.EncryptFRMPayload(node.AppSKey); err != nil {
		panic(err)
	}

	payload := lorawan.NewPHYPayload(uplink)
	payload.MHDR = lorawan.MHDR{
		MType: lorawan.ConfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	payload.MACPayload = macPayload

	if err := payload.SetMIC(node.NwkSKey); err != nil {
		panic(err)
	}

	bytes, err := payload.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return bytes
}
