// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package band

import "time"

// BandLimits_TTN are applied to unknown bands
var BandLimits_TTN = BandLimits{
	SubBandLimits{
		appliesTo: func(_ uint64) bool { return true },
		limits: UtilizationLimits{
			time.Hour: 1.0,
		},
	},
}

// BandLimits_EU_863_870 is the duty cycle configuration for the EU_863_870 band
var BandLimits_EU_863_870 = append(BandLimits_TTN,
	SubBandLimits{
		// g 863.0 – 868.0 MHz 1%
		appliesTo: func(freq uint64) bool { return freq >= 863000000 && freq < 868000000 },
		limits: UtilizationLimits{
			time.Hour: 0.01,
		},
	},
	SubBandLimits{
		// g1 868.0 – 868.6 MHz 1%
		appliesTo: func(freq uint64) bool { return freq >= 868000000 && freq < 868600000 },
		limits: UtilizationLimits{
			time.Hour: 0.01,
		},
	},
	SubBandLimits{
		// g2 868.7 – 869.2 MHz 0.1%
		appliesTo: func(freq uint64) bool { return freq >= 868700000 && freq < 869200000 },
		limits: UtilizationLimits{
			time.Hour: 0.001,
		},
	},
	SubBandLimits{
		// g3 869.4 – 869.65 MHz 10%
		appliesTo: func(freq uint64) bool { return freq >= 869400000 && freq < 869650000 },
		limits: UtilizationLimits{
			time.Hour: 0.1,
		},
	},
	SubBandLimits{
		// g4 869.7 – 870.0 MHz 1%
		appliesTo: func(freq uint64) bool { return freq >= 869700000 && freq < 870000000 },
		limits: UtilizationLimits{
			time.Hour: 0.01,
		},
	},
	SubBandLimits{
		// LoRaWAN Join Channels
		appliesTo: func(freq uint64) bool { return freq == 867100000 || freq == 867300000 || freq == 867500000 },
		limits: UtilizationLimits{
			time.Minute: 0.01,
		},
	},
)

// BandLimits_CN_779_787 is the duty cycle configuration for the CN_779_787 band
var BandLimits_CN_779_787 = append(BandLimits_TTN,
	SubBandLimits{
		appliesTo: func(_ uint64) bool { return true },
		limits: UtilizationLimits{
			5 * time.Minute: 0.01,
		},
	},
)

// BandLimits_EU_433 is the duty cycle configuration for the EU_433 band
var BandLimits_EU_433 = append(BandLimits_TTN,
	SubBandLimits{
		appliesTo: func(_ uint64) bool { return true },
		limits: UtilizationLimits{
			5 * time.Minute: 0.01,
		},
	},
)

// BandLimits_AS_923 is the duty cycle configuration for the AS_923 band
var BandLimits_AS_923 = append(BandLimits_TTN,
	SubBandLimits{
		appliesTo: func(_ uint64) bool { return true },
		limits: UtilizationLimits{
			5 * time.Minute: 0.01,
		},
	},
)
