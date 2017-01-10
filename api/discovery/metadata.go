// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import "github.com/TheThingsNetwork/ttn/core/types"

// AppIDs that are handled by this component
func (a *Announcement) AppIDs() (appIDs []string) {
	for _, meta := range a.Metadata {
		if appID := meta.GetAppId(); appID != "" {
			appIDs = append(appIDs, appID)
		}
	}
	return
}

// DevAddrPrefixes that are handled by this component
func (a *Announcement) DevAddrPrefixes() (prefixes []types.DevAddrPrefix) {
	for _, meta := range a.Metadata {
		if prefixBytes := meta.GetDevAddrPrefix(); prefixBytes != nil {
			prefix := new(types.DevAddrPrefix)
			if err := prefix.Unmarshal(prefixBytes); err != nil {
				continue
			}
			prefixes = append(prefixes, *prefix)
		}
	}
	return
}

// AppEUIs that are handled by this component
func (a *Announcement) AppEUIs() (euis []types.AppEUI) {
	for _, meta := range a.Metadata {
		if euiBytes := meta.GetAppEui(); euiBytes != nil {
			eui := new(types.AppEUI)
			if err := eui.Unmarshal(euiBytes); err != nil {
				continue
			}
			euis = append(euis, *eui)
		}
	}
	return
}
