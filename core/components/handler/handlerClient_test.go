// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"golang.org/x/net/context"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient("0.0.0.0:12345")
	CheckErrors(t, nil, err)
}

func TestListDevices(t *testing.T) {
	{
		Desc(t, "Valid request, no issue")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		st.OutReadAll.Entries = []devEntry{
			{
				DevEUI:   []byte{1, 2, 1, 2, 1, 2, 1, 2},
				DevAddr:  []byte{14, 14, 14, 14},
				AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
				NwkSKey:  [16]byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
				AppSKey:  [16]byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
				FCntDown: 14,
				AppKey:   &[16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			},
			{
				DevAddr:  []byte{42, 42, 42, 42},
				AppEUI:   []byte{8, 7, 6, 5, 4, 3, 2, 1},
				NwkSKey:  [16]byte{1, 2, 3, 6, 1, 2, 3, 6, 1, 2, 3, 6, 1, 2, 3, 6},
				AppSKey:  [16]byte{48, 2, 48, 2, 48, 2, 48, 2, 48, 2, 48, 2, 48, 2, 48, 2},
				FCntDown: 5,
			},
		}
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.ListDevicesHandlerReq{
			Token:  "==OAuth==Token==",
			AppEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr *string
		var wantBrkCall = &core.ValidateTokenBrokerReq{
			Token:  req.Token,
			AppEUI: req.AppEUI,
		}
		var wantRes = &core.ListDevicesHandlerRes{
			OTAA: []*core.HandlerOTAADevice{
				&core.HandlerOTAADevice{
					DevEUI:   st.OutReadAll.Entries[0].DevEUI,
					DevAddr:  st.OutReadAll.Entries[0].DevAddr,
					NwkSKey:  st.OutReadAll.Entries[0].NwkSKey[:],
					AppSKey:  st.OutReadAll.Entries[0].AppSKey[:],
					AppKey:   st.OutReadAll.Entries[0].AppKey[:],
					FCntDown: 14,
					FCntUp:   0,
				},
			},
			ABP: []*core.HandlerABPDevice{
				&core.HandlerABPDevice{
					DevAddr:  st.OutReadAll.Entries[1].DevAddr,
					NwkSKey:  st.OutReadAll.Entries[1].NwkSKey[:],
					AppSKey:  st.OutReadAll.Entries[1].AppSKey[:],
					FCntDown: 5,
					FCntUp:   0,
				},
			},
		}

		// Operate
		res, err := h.ListDevices(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InValidateToken.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Valid request | readAll fails")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		st.Failures["readAll"] = errors.New(errors.Operational, "Mock Error")
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.ListDevicesHandlerReq{
			Token:  "==OAuth==Token==",
			AppEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrOperational
		var wantBrkCall = &core.ValidateTokenBrokerReq{
			Token:  req.Token,
			AppEUI: req.AppEUI,
		}
		var wantRes = new(core.ListDevicesHandlerRes)

		// Operate
		res, err := h.ListDevices(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InValidateToken.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Valid request | broker fails")

		// Build
		br := mocks.NewAuthBrokerClient()
		br.Failures["ValidateToken"] = errors.New(errors.Operational, "Mock Error")
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.ListDevicesHandlerReq{
			Token:  "==OAuth==Token==",
			AppEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrOperational
		var wantBrkCall = &core.ValidateTokenBrokerReq{
			Token:  req.Token,
			AppEUI: req.AppEUI,
		}
		var wantRes = new(core.ListDevicesHandlerRes)

		// Operate
		res, err := h.ListDevices(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InValidateToken.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Invalid AppEUI")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.ListDevicesHandlerReq{
			Token:  "==OAuth==Token==",
			AppEUI: []byte{1, 2, 3, 4, 5},
		}

		// Expect
		var wantErr = ErrStructural
		var wantBrkCall *core.ValidateTokenBrokerReq
		var wantRes = new(core.ListDevicesHandlerRes)

		// Operate
		res, err := h.ListDevices(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InValidateToken.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}
}

func TestUpsertABP(t *testing.T) {
	{
		Desc(t, "Valid request, no issue")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertABPHandlerReq{
			Token:   "==OAuth==Token==",
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevAddr: []byte{14, 14, 14, 14},
			NwkSKey: []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			AppSKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr *string
		var wantBrkCall = &core.UpsertABPBrokerReq{
			Token:      req.Token,
			AppEUI:     req.AppEUI,
			DevAddr:    []byte{14, 14, 14, 14},
			NwkSKey:    []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			NetAddress: h.(*component).PrivateNetAddr,
		}
		var wantRes = new(core.UpsertABPHandlerRes)

		// Operate
		res, err := h.UpsertABP(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InUpsertABP.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Valid request | storage fails")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		st.Failures["upsert"] = errors.New(errors.Operational, "Mock Error")
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertABPHandlerReq{
			Token:   "==OAuth==Token==",
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevAddr: []byte{14, 14, 14, 14},
			NwkSKey: []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			AppSKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrOperational
		var wantBrkCall = &core.UpsertABPBrokerReq{
			Token:      req.Token,
			AppEUI:     req.AppEUI,
			DevAddr:    []byte{14, 14, 14, 14},
			NwkSKey:    []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			NetAddress: h.(*component).PrivateNetAddr,
		}
		var wantRes = new(core.UpsertABPHandlerRes)

		// Operate
		res, err := h.UpsertABP(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InUpsertABP.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Valid request | broker fails")

		// Build
		br := mocks.NewAuthBrokerClient()
		br.Failures["UpsertABP"] = errors.New(errors.Operational, "Mock Error")
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertABPHandlerReq{
			Token:   "==OAuth==Token==",
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevAddr: []byte{14, 14, 14, 14},
			NwkSKey: []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			AppSKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrOperational
		var wantBrkCall = &core.UpsertABPBrokerReq{
			Token:      req.Token,
			AppEUI:     req.AppEUI,
			DevAddr:    []byte{14, 14, 14, 14},
			NwkSKey:    []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			NetAddress: h.(*component).PrivateNetAddr,
		}
		var wantRes = new(core.UpsertABPHandlerRes)

		// Operate
		res, err := h.UpsertABP(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InUpsertABP.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Invalid request | DevAddr invalid")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertABPHandlerReq{
			Token:   "==OAuth==Token==",
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevAddr: []byte{14, 14, 14, 14, 14, 14},
			NwkSKey: []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			AppSKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrStructural
		var wantBrkCall *core.UpsertABPBrokerReq
		var wantRes = new(core.UpsertABPHandlerRes)

		// Operate
		res, err := h.UpsertABP(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InUpsertABP.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Invalid request | AppEUI invalid")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertABPHandlerReq{
			Token:   "==OAuth==Token==",
			AppEUI:  []byte{1, 2, 3, 4, 5, 6},
			DevAddr: []byte{14, 14, 14, 14},
			NwkSKey: []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			AppSKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrStructural
		var wantBrkCall *core.UpsertABPBrokerReq
		var wantRes = new(core.UpsertABPHandlerRes)

		// Operate
		res, err := h.UpsertABP(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InUpsertABP.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Invalid request | NwkSKey invalid")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertABPHandlerReq{
			Token:   "==OAuth==Token==",
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevAddr: []byte{14, 14, 14, 14},
			NwkSKey: nil,
			AppSKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrStructural
		var wantBrkCall *core.UpsertABPBrokerReq
		var wantRes = new(core.UpsertABPHandlerRes)

		// Operate
		res, err := h.UpsertABP(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InUpsertABP.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Invalid request | AppSKey invalid")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertABPHandlerReq{
			Token:   "==OAuth==Token==",
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevAddr: []byte{14, 14, 14, 14},
			NwkSKey: []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			AppSKey: []byte{1, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrStructural
		var wantBrkCall *core.UpsertABPBrokerReq
		var wantRes = new(core.UpsertABPHandlerRes)

		// Operate
		res, err := h.UpsertABP(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InUpsertABP.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}
}

func TestUpsertOTAA(t *testing.T) {
	{
		Desc(t, "Valid request, no issue")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertOTAAHandlerReq{
			Token:  "==OAuth==Token==",
			AppEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI: []byte{14, 14, 14, 14, 14, 14, 14, 14},
			AppKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr *string
		var wantBrkCall = &core.ValidateOTAABrokerReq{
			Token:      req.Token,
			AppEUI:     req.AppEUI,
			NetAddress: h.(*component).PrivateNetAddr,
		}
		var wantRes = new(core.UpsertOTAAHandlerRes)

		// Operate
		res, err := h.UpsertOTAA(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InValidateOTAA.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Valid request | storage fails")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		st.Failures["upsert"] = errors.New(errors.Operational, "Mock Error")
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertOTAAHandlerReq{
			Token:  "==OAuth==Token==",
			AppEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI: []byte{14, 14, 14, 14, 14, 14, 14, 14},
			AppKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrOperational
		var wantBrkCall = &core.ValidateOTAABrokerReq{
			Token:      req.Token,
			AppEUI:     req.AppEUI,
			NetAddress: h.(*component).PrivateNetAddr,
		}
		var wantRes = new(core.UpsertOTAAHandlerRes)

		// Operate
		res, err := h.UpsertOTAA(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InValidateOTAA.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Valid request | broker fails")

		// Build
		br := mocks.NewAuthBrokerClient()
		br.Failures["ValidateOTAA"] = errors.New(errors.Operational, "Mock Error")
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertOTAAHandlerReq{
			Token:  "==OAuth==Token==",
			AppEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI: []byte{14, 14, 14, 14, 14, 14, 14, 14},
			AppKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrOperational
		var wantBrkCall = &core.ValidateOTAABrokerReq{
			Token:      req.Token,
			AppEUI:     req.AppEUI,
			NetAddress: h.(*component).PrivateNetAddr,
		}
		var wantRes = new(core.UpsertOTAAHandlerRes)

		// Operate
		res, err := h.UpsertOTAA(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InValidateOTAA.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Invalid request | DevEUI invalid")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertOTAAHandlerReq{
			Token:  "==OAuth==Token==",
			AppEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI: []byte{14, 14, 14, 14, 14, 14},
			AppKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrStructural
		var wantBrkCall *core.ValidateOTAABrokerReq
		var wantRes = new(core.UpsertOTAAHandlerRes)

		// Operate
		res, err := h.UpsertOTAA(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InValidateOTAA.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Invalid request | AppEUI invalid")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertOTAAHandlerReq{
			Token:  "==OAuth==Token==",
			AppEUI: []byte{1, 2, 3, 4, 5, 6},
			DevEUI: []byte{14, 14, 14, 14, 14, 14, 14, 14},
			AppKey: []byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrStructural
		var wantBrkCall *core.ValidateOTAABrokerReq
		var wantRes = new(core.UpsertOTAAHandlerRes)

		// Operate
		res, err := h.UpsertOTAA(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InValidateOTAA.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}

	// --------------------

	{
		Desc(t, "Invalid request | AppKey invalid")

		// Build
		br := mocks.NewAuthBrokerClient()
		st := NewMockDevStorage()
		h := New(
			Components{
				Ctx:        GetLogger(t, "Handler"),
				Broker:     br,
				DevStorage: st,
			}, Options{
				PublicNetAddr:  "NetAddr",
				PrivateNetAddr: "PrivNetAddr",
			})
		req := &core.UpsertOTAAHandlerReq{
			Token:  "==OAuth==Token==",
			AppEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI: []byte{14, 14, 14, 14, 14, 14, 14, 14},
			AppKey: []byte{1, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		// Expect
		var wantErr = ErrStructural
		var wantBrkCall *core.ValidateOTAABrokerReq
		var wantRes = new(core.UpsertOTAAHandlerRes)

		// Operate
		res, err := h.UpsertOTAA(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantBrkCall, br.InValidateOTAA.Req, "Broker Calls")
		Check(t, wantRes, res, "Handler responses")
	}
}
