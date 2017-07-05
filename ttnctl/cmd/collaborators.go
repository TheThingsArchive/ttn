// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/go-account-lib/rights"
	"github.com/TheThingsNetwork/ttn/core/types"
)

func printCollaborators(collaborators []account.Collaborator) {
	for _, collaborator := range collaborators {
		fmt.Println("       - Username:", collaborator.Username)
		fmt.Println("         Rights:", joinRights(collaborator.Rights, ", "))
	}
}

var applicationRights = []types.Right{
	rights.AppSettings,
	rights.AppCollaborators,
	rights.AppDelete,
	rights.Devices,
}
var gatewayRights = []types.Right{
	rights.GatewaySettings,
	rights.GatewayCollaborators,
	rights.GatewayDelete,
	rights.GatewayLocation,
	rights.GatewayStatus,
	rights.GatewayOwner,
	rights.GatewayMessages,
}
var componentsRights = []types.Right{
	rights.ComponentSettings,
	rights.ComponentDelete,
	rights.ComponentCollaborators,
}

func validRight(available []types.Right, right types.Right) bool {
	for _, available := range available {
		if right == available {
			return true
		}
	}
	return false
}

func joinRights(rights []types.Right, sep string) string {
	rightStrings := make([]string, len(rights))
	for i, right := range rights {
		rightStrings[i] = string(right)
	}
	return strings.Join(rightStrings, sep)
}
