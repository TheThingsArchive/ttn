package main

import (
	"fmt"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/semtech"
	"net"
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
			packet, err := core.ConvertRXPK(pkt.Payload.RXPK[0])
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(packet)
		}
	}()

	fmt.Println("Listening on 33000")
	<-make(chan bool)
}
