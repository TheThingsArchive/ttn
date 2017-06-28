// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
)

var componentsCollaboratorsCmd = &cobra.Command{
	Use:   "collaborators",
	Short: "Manage collaborators of a network component.",
	Long:  `components collaborators can be used to manage the collaborators of a network component.`,
}

func init() {
	componentsCmd.AddCommand(componentsCollaboratorsCmd)
}
