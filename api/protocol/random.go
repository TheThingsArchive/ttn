// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package protocol

import "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"

// RandomTxConfiguration returns randomly generated TxConfiguration.
// Used for testing.
func RandomTxConfiguration() *TxConfiguration {
	return RandomLorawanTxConfiguration()
}

// RandomRxMetadata returns randomly generated RxMetadata.
// Used for testing.
func RandomRxMetadata() *RxMetadata {
	return RandomLorawanRxMetadata()
}

// RandomLorawanRxMetadata returns randomly generated LorawanRxMetadata.
// Used for testing.
func RandomLorawanRxMetadata(modulation ...lorawan.Modulation) *RxMetadata {
	return &RxMetadata{
		Protocol: &RxMetadata_Lorawan{
			Lorawan: lorawan.RandomMetadata(modulation...),
		},
	}
}

// RandomLorawanTxConfiguration returns randomly generated LorawanTxConfiguration.
// Used for testing.
func RandomLorawanTxConfiguration(modulation ...lorawan.Modulation) *TxConfiguration {
	return &TxConfiguration{
		Protocol: &TxConfiguration_Lorawan{
			Lorawan: lorawan.RandomTxConfiguration(modulation...),
		},
	}
}
