// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

// MockHandler implements the udp.Handler interface.
type MockHandler struct {
	OutMsgUDP []byte
	OutMsgReq []byte
	InMsg     MsgUDP
	InChresp  []byte
}

// Handle implements the udp.Handler interface
func (h *MockHandler) Handle(conn chan<- MsgUDP, next chan<- MsgReq, msg MsgUDP) error {
	h.InMsg = msg
	if h.OutMsgReq != nil {
		chresp := make(chan MsgRes)
		next <- MsgReq{
			Data:   h.OutMsgReq,
			Chresp: chresp,
		}
		h.InChresp = <-chresp
	}

	if h.OutMsgUDP != nil {
		conn <- MsgUDP{
			Data: h.OutMsgUDP,
			Addr: msg.Addr,
		}
	}
	return nil
}
