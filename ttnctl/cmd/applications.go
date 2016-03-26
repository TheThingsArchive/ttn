// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// applicationsCmd represents the applications command
var applicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "Show applications",
	Long:  `ttnctl applications retrieves your applications of the logged on user.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("applications called")
	},
}

func init() {
	RootCmd.AddCommand(applicationsCmd)
}
