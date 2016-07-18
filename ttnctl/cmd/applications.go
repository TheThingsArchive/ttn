// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import "github.com/spf13/cobra"

// applicationsCmd is the entrypoint for handlerctl
var applicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "Manage applications",
	Long:  `ttnctl applications can be used to manage applications.`,
}

func init() {
	RootCmd.AddCommand(applicationsCmd)
}
