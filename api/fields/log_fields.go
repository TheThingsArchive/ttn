// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package fields

import (
	"strings"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// Debug mode
var Debug bool

type hasId interface {
	GetId() string
}

type hasServiceName interface {
	GetServiceName() string
}

func fillDiscoveryFields(m interface{}, f log.Fields) {
	if m, ok := m.(hasId); ok {
		if v := m.GetId(); v != "" {
			f["ID"] = v
		}
	}
	if m, ok := m.(hasServiceName); ok {
		if v := m.GetServiceName(); v != "" {
			f["ServiceName"] = v
		}
	}
}

type hasDevId interface {
	GetDevId() string
}

type hasAppId interface {
	GetAppId() string
}

type hasDevEui interface {
	GetDevEui() *types.DevEUI
}

type hasAppEui interface {
	GetAppEui() *types.AppEUI
}

type hasDevAddr interface {
	GetDevAddr() *types.DevAddr
}

func fillIdentifiers(m interface{}, f log.Fields) {
	if m, ok := m.(hasDevEui); ok {
		if v := m.GetDevEui(); v != nil {
			f["DevEUI"] = v
		}
	}
	if m, ok := m.(hasAppEui); ok {
		if v := m.GetAppEui(); v != nil {
			f["AppEUI"] = v
		}
	}
	if m, ok := m.(hasDevId); ok {
		if v := m.GetDevId(); v != "" {
			f["DevID"] = v
		}
	}
	if m, ok := m.(hasAppId); ok {
		if v := m.GetAppId(); v != "" {
			f["AppID"] = v
		}
	}
	if m, ok := m.(hasDevAddr); ok {
		if v := m.GetDevAddr(); v != nil {
			f["DevAddr"] = v
		}
	}
}

type hasProtocolMetadata interface {
	GetProtocolMetadata() *protocol.RxMetadata
}

type hasProtocolConfiguration interface {
	GetProtocolConfiguration() *protocol.TxConfiguration
}

type hasProtocolConfig interface {
	GetProtocolConfig() *protocol.TxConfiguration
}

func fillProtocolConfig(cfg *protocol.TxConfiguration, f log.Fields) {
	if lorawan := cfg.GetLorawan(); lorawan != nil {
		if v := lorawan.Modulation.String(); v != "" {
			f["Modulation"] = v
		}
		if v := lorawan.DataRate; v != "" {
			f["DataRate"] = v
		}
		if v := lorawan.BitRate; v != 0 {
			f["BitRate"] = v
		}
		if v := lorawan.CodingRate; v != "" {
			f["CodingRate"] = v
		}
	}
}

func fillProtocol(m interface{}, f log.Fields) {
	if m, ok := m.(hasProtocolMetadata); ok {
		if meta := m.GetProtocolMetadata(); meta != nil {
			if lorawan := meta.GetLorawan(); lorawan != nil {
				if v := lorawan.Modulation.String(); v != "" {
					f["Modulation"] = v
				}
				if v := lorawan.DataRate; v != "" {
					f["DataRate"] = v
				}
				if v := lorawan.BitRate; v != 0 {
					f["BitRate"] = v
				}
				if v := lorawan.CodingRate; v != "" {
					f["CodingRate"] = v
				}
			}
		}
	}
	if m, ok := m.(hasProtocolConfiguration); ok {
		if cfg := m.GetProtocolConfiguration(); cfg != nil {
			fillProtocolConfig(cfg, f)
		}
	}
	if m, ok := m.(hasProtocolConfig); ok {
		if cfg := m.GetProtocolConfig(); cfg != nil {
			fillProtocolConfig(cfg, f)
		}
	}
}

type hasGatewayConfiguration interface {
	GetGatewayConfiguration() *gateway.TxConfiguration
}

type hasGatewayConfig interface {
	GetGatewayConfig() *gateway.TxConfiguration
}

type hasGatewayMetadata interface {
	GetGatewayMetadata() *gateway.RxMetadata
}

type hasMoreGatewayMetadata interface {
	GetGatewayMetadata() []*gateway.RxMetadata
}

func fillGatewayConfig(cfg *gateway.TxConfiguration, f log.Fields) {
	if v := cfg.Frequency; v != 0 {
		f["Frequency"] = v
	}
	if v := cfg.Power; v != 0 {
		f["Power"] = v
	}
}

func fillGateway(m interface{}, f log.Fields) {
	if m, ok := m.(hasGatewayMetadata); ok {
		if meta := m.GetGatewayMetadata(); meta != nil {
			if v := meta.GatewayId; v != "" {
				f["GatewayID"] = v
			}
			if v := meta.Frequency; v != 0 {
				f["Frequency"] = v
			}
			if v := meta.Rssi; v != 0 {
				f["RSSI"] = v
			}
			if v := meta.Snr; v != 0 {
				f["SNR"] = v
			}
		}
	}
	if m, ok := m.(hasMoreGatewayMetadata); ok {
		if meta := m.GetGatewayMetadata(); meta != nil {
			f["NumGateways"] = len(meta)
		}
	}
	if m, ok := m.(hasGatewayConfiguration); ok {
		if cfg := m.GetGatewayConfiguration(); cfg != nil {
			fillGatewayConfig(cfg, f)
		}
	}
	if m, ok := m.(hasGatewayConfig); ok {
		if cfg := m.GetGatewayConfig(); cfg != nil {
			fillGatewayConfig(cfg, f)
		}
	}
}

type hasMessage interface {
	GetMessage() *protocol.Message
}

type hasPayload interface {
	GetPayload() []byte
}

func fillMessage(m interface{}, f log.Fields) {
	var payload []byte
	if m, ok := m.(hasPayload); ok {
		payload = m.GetPayload()
		f["PayloadSize"] = len(payload)
	}
	if m, ok := m.(hasMessage); ok {
		m := m.GetMessage()
		if m == nil && Debug {
			if msg, err := lorawan.MessageFromPHYPayloadBytes(payload); err == nil {
				m = new(protocol.Message)
				m.Protocol = &protocol.Message_Lorawan{Lorawan: &msg}
			}
		}
		if m != nil {
			if lorawan := m.GetLorawan(); lorawan != nil {
				if mac := lorawan.GetMacPayload(); mac != nil {
					if v := mac.DevAddr; !v.IsEmpty() {
						f["DevAddr"] = v
					}
					if v := len(mac.FrmPayload); v != 0 {
						f["AppPayloadSize"] = v
					}
					if v := mac.FPort; v != 0 {
						f["Port"] = v
					}
					if v := mac.FCnt; v != 0 {
						f["Counter"] = v
					}
					fillMAC(mac, f)
				}
				if join := lorawan.GetJoinRequestPayload(); join != nil {
					f["AppEUI"] = join.AppEui
					f["DevEUI"] = join.DevEui
				}
			}
		}
	}
}

func fillMAC(m *lorawan.MACPayload, f log.Fields) {
	var mac []string
	if m.Ack {
		mac = append(mac, "Ack")
	}
	if m.Adr {
		mac = append(mac, "Adr")
	}
	if m.FPending {
		mac = append(mac, "FPending")
	}
	if m.AdrAckReq {
		mac = append(mac, "AdrAckReq")
	}
	for _, m := range m.FOpts {
		switch m.Cid {
		case 0x02:
			mac = append(mac, "LinkCheck")
		case 0x03:
			mac = append(mac, "LinkADR")
		case 0x04:
			mac = append(mac, "DutyCycle")
		case 0x05:
			mac = append(mac, "RXParamSetup")
		case 0x06:
			mac = append(mac, "DevStatus")
		case 0x07:
			mac = append(mac, "NewChannel")
		case 0x08:
			mac = append(mac, "RXTimingSetup")
		case 0x09:
			mac = append(mac, "TXParamSetup")
		case 0x0A:
			mac = append(mac, "DLChannel")
		}
	}
	f["MAC"] = strings.Join(mac, ",")
}

// Get a number of log fields for a message, if we're able to extract them
func Get(m interface{}) log.Fields {
	fields := log.Fields{}
	fillDiscoveryFields(m, fields)
	fillIdentifiers(m, fields)
	fillGateway(m, fields)
	fillProtocol(m, fields)
	fillMessage(m, fields)
	return fields
}
