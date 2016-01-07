// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package brk_hdl_http

import (
	"bytes"
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
	"github.com/thethingsnetwork/core/utils/log"
	. "github.com/thethingsnetwork/core/utils/testing"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestNewAdapter(t *testing.T) {
	tests := []struct {
		Port      uint
		WantError error
	}{
		{3000, nil},
		{0, ErrInvalidPort},
	}

	for _, test := range tests {
		_, err := NewAdapter(test.Port)
		checkErrors(t, test.WantError, err)
	}
}

type nextRegistrationTest struct {
	AppId      string
	AppUrl     string
	DevAddr    string
	NwsKey     string
	WantResult nextRegistrationResult
}

type nextRegistrationResult struct {
	Config  *core.Registration
	AckNack core.AckNacker
	Error   error
}

func TestNextRegistration(t *testing.T) {
	tests := []nextRegistrationTest{
		// Valid device address
		{
			AppId:   "appid",
			AppUrl:  "myhandler.com:3000",
			NwsKey:  "00112233445566778899aabbccddeeff",
			DevAddr: "14aab0a4",
			WantResult: nextRegistrationResult{
				Config: &core.Registration{
					DevAddr: lorawan.DevAddr([4]byte{14, 0xaa, 0xb0, 0xa4}),
					Handler: core.Recipient{Id: "appid", Address: "myhandler.com:3000"},
					NwsKey:  lorawan.AES128Key([16]byte{00, 11, 22, 33, 44, 55, 66, 77, 88, 99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}),
				},
				Error: nil,
			},
		},
		// Invalid device address
		{
			AppId:   "appid",
			AppUrl:  "myhandler.com:3000",
			NwsKey:  "00112233445566778899aabbccddeeff",
			DevAddr: "INVALID",
			WantResult: nextRegistrationResult{
				Config: nil,
				Error:  nil,
			},
		},
		// Invalid nwskey address
		{
			AppId:   "appid",
			AppUrl:  "myhandler.com:3000",
			NwsKey:  "00112233445566778899af",
			DevAddr: "14aaab0a4",
			WantResult: nextRegistrationResult{
				Config: nil,
				Error:  nil,
			},
		},
	}

	adapter, err := NewAdapter(3001)
	adapter.logger = log.TestLogger{Tag: "BRK_HDL_ADAPTER", T: t}
	client := &client{
		adapter: "0.0.0.0:3001",
		c:       http.Client{},
		logger:  log.TestLogger{Tag: "http client", T: t},
	}
	<-time.After(time.Millisecond * 200)
	if err != nil {
		panic(err)
	}

	for _, test := range tests {
		client.send(test.AppId, test.AppUrl, test.DevAddr, test.NwsKey)
		res := make(chan nextRegistrationResult)
		go func() {
			config, an, err := adapter.NextRegistration()
			res <- nextRegistrationResult{&config, an, err}
		}()

		select {
		case result := <-res:
			checkRegistrationResult(t, test.WantResult, result)
		case <-time.After(time.Millisecond * 250):
			checkRegistrationResult(t, test.WantResult, nextRegistrationResult{})
		}
	}
}

func checkErrors(t *testing.T, want error, got error) bool {
	if want == got {
		Ok(t, "Check errors")
		return true
	}
	Ko(t, "Expected error to be {%v} but got {%v}", want, got)
	return false
}

func checkRegistrationResult(t *testing.T, want nextRegistrationResult, got nextRegistrationResult) bool {
	if !checkErrors(t, want.Error, got.Error) {
		return false
	}

	if want.Config == nil {
		if got.Error == nil || got.AckNack != nil {
			Ko(t, "Was expecting no result but got %v", got.Config)
			return false
		}
		Ok(t, "Check registration result")
		return true
	}

	if !reflect.DeepEqual(*want.Config, *got.Config) {
		Ko(t, "Received configuration doesn't match expectations\nWant: %v\nGot:  %v", *want.Config, *got.Config)
		return false
	}

	if want.AckNack == nil {
		Ko(t, "Received configuration with a nil AckNacker")
		return false
	}

	Ok(t, "Check registration result")
	return true
}

type client struct {
	c       http.Client
	logger  log.Logger
	adapter string
}

func (c *client) send(appId, appUrl, devAddr, nwsKey string) {
	c.logger.Log("send request to %s", c.adapter)
	buf := new(bytes.Buffer)
	if _, err := buf.WriteString(fmt.Sprintf(`{"app_id":"%s","app_url":"%s","nws_key":"%s"}`, appId, appUrl, nwsKey)); err != nil {
		panic(err)
	}
	resp, err := c.c.Post(fmt.Sprintf("http://%s/end-device/%s", c.adapter, devAddr), "application/json", buf)
	if err != nil {
		panic(err)
	}
	c.logger.Log("response code: %d", resp.StatusCode)
}
