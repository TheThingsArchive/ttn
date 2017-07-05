// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
)

var applicationsCollaboratorsCmd = &cobra.Command{
	Use:   "collaborators",
	Short: "Manage collaborators of an application.",
	Long:  `applications collaborators can be used to manage the collaborators of an application.`,
}

func init() {
	applicationsCmd.AddCommand(applicationsCollaboratorsCmd)
}
