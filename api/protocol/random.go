package protocol

import "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"

func RandomTxConfiguration() *TxConfiguration {
	return RandomLorawanTxConfiguration()
}

func RandomRxMetadata() *RxMetadata {
	return RandomLorawanRxMetadata()
}

func RandomLorawanRxMetadata(modulation ...lorawan.Modulation) *RxMetadata {
	return &RxMetadata{
		Protocol: &RxMetadata_Lorawan{
			Lorawan: lorawan.RandomMetadata(modulation...),
		},
	}
}

func RandomLorawanTxConfiguration(modulation ...lorawan.Modulation) *TxConfiguration {
	return &TxConfiguration{
		Protocol: &TxConfiguration_Lorawan{
			Lorawan: lorawan.RandomTxConfiguration(modulation...),
		},
	}
}
