// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

// PayloadUnmarshaler unmarshals the Payload to a Message
type PayloadUnmarshaler interface {
	UnmarshalPayload() error
}
