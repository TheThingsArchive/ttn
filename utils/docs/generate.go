// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package docs

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].CommandPath() < s[j].CommandPath() }

// Generate prints API docs for a command
func Generate(cmd *cobra.Command) string {
	buf := new(bytes.Buffer)

	cmds := genCmdList(cmd)
	sort.Sort(byName(cmds))
	for _, cmd := range cmds {
		if len(strings.Split(cmd.CommandPath(), " ")) == 1 {
			fmt.Fprint(buf, "**Options**\n\n")
			printOptions(buf, cmd)
			fmt.Fprintln(buf)
			continue
		}

		depth := len(strings.Split(cmd.CommandPath(), " "))

		fmt.Fprint(buf, header(depth, cmd.CommandPath()))

		fmt.Fprint(buf, cmd.Long, "\n\n")

		if cmd.Runnable() {
			fmt.Fprint(buf, "**Usage:** ", "`", cmd.UseLine(), "`", "\n\n")
		}

		if cmd.HasLocalFlags() || cmd.HasPersistentFlags() {
			fmt.Fprint(buf, "**Options**\n\n")
			printOptions(buf, cmd)
		}

		if cmd.Example != "" {
			fmt.Fprint(buf, "**Example**\n\n")
			fmt.Fprint(buf, "```", "\n", cmd.Example, "```", "\n\n")
		}
	}

	return buf.String()
}

func genCmdList(cmd *cobra.Command) (cmds []*cobra.Command) {
	cmds = append(cmds, cmd)
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		cmds = append(cmds, genCmdList(c)...)
	}
	return cmds
}

func header(depth int, header string) string {
	return fmt.Sprint(strings.Repeat("#", depth), " ", header, "\n\n")
}

func printOptions(w io.Writer, cmd *cobra.Command) {
	fmt.Fprintln(w, "```")
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(w)
	flags.PrintDefaults()
	fmt.Fprintln(w, "```")
	fmt.Fprintln(w, "")
}
