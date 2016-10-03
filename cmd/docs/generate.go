package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/TheThingsNetwork/ttn/cmd"
	"github.com/spf13/cobra"
)

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].CommandPath() < s[j].CommandPath() }

func main() {
	cmds := genCmdList(cmd.RootCmd)
	sort.Sort(byName(cmds))
	fmt.Println(`# API Reference

The Things Network's backend servers.
`)
	for _, cmd := range cmds {
		if cmd.CommandPath() == "ttn" {
			fmt.Print("**Options**\n\n")
			printOptions(cmd)
			fmt.Println()
			continue
		}

		depth := len(strings.Split(cmd.CommandPath(), " "))

		printHeader(depth, cmd.CommandPath())

		fmt.Print(cmd.Long, "\n\n")

		if cmd.Runnable() {
			fmt.Print("**Usage:** ", "`", cmd.UseLine(), "`", "\n\n")
		}

		if cmd.HasLocalFlags() || cmd.HasPersistentFlags() {
			fmt.Print("**Options**\n\n")
			printOptions(cmd)
		}

		if cmd.Example != "" {
			fmt.Print("**Example**\n\n")
			fmt.Print("```", "\n", cmd.Example, "```", "\n\n")
		}
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

func printHeader(depth int, header string) {
	fmt.Print(strings.Repeat("#", depth), " ", header, "\n\n")
}

func printOptions(cmd *cobra.Command) {
	fmt.Println("```")
	var b bytes.Buffer
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(&b)
	flags.PrintDefaults()
	fmt.Print(b.String())
	fmt.Println("```")
}
