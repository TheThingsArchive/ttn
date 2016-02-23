// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding"
)

// MqttRecipient describes recipient manipulated by the mqtt adapter
type MqttRecipient interface {
	encoding.BinaryMarshaler
	TopicUp() string
	TopicDown() string
}
