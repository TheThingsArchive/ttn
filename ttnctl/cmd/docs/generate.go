package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/TheThingsNetwork/ttn/ttnctl/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].CommandPath() < s[j].CommandPath() }

func main() {
	cmds := genCmdList(cmd.RootCmd)
	sort.Sort(byName(cmds))
	for _, cmd := range cmds {
		var buf bytes.Buffer
		doc.GenMarkdownCustom(cmd, &buf, func(s string) string {
			return "#" + strings.TrimSuffix(s, ".md")
		})
		cleaned := strings.Split(buf.String(), "### SEE ALSO")[0]
		fmt.Println(cleaned)
	}
}

func genCmdList(cmd *cobra.Command) (cmds []*cobra.Command) {
	cmds = append(cmds, cmd)
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsHelpCommand() {
			continue
		}
		cmds = append(cmds, genCmdList(c)...)
	}
	return cmds
}
