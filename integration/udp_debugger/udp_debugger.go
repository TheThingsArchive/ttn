// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net"

	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:33000")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			fmt.Printf("\n*******************************\n")
			buf := make([]byte, 512)
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Msg from", addr)
			pkt := new(semtech.Packet)
			if err := pkt.UnmarshalBinary(buf[:n]); err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Printf("Received %x from %x with token %x\n", pkt.Identifier, pkt.GatewayId, pkt.Token)

			if pkt.Payload == nil || len(pkt.Payload.RXPK) < 1 {
				fmt.Println("Unexpected packet payload")
				continue
			}
			fmt.Printf(pointer.DumpPStruct(pkt.Payload.RXPK[0], true))
		}
	}()

	fmt.Println("Listening on 33000")
	<-make(chan bool)
}
