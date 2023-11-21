package main

import (
	"fmt"
	"io"

	"github.com/mkobetic/coin"
)

func init() {
	(&cmdVersion{}).newCommand("version", "ver", "v")
}

type cmdVersion struct {
	flagsWithUsage
}

func (*cmdVersion) newCommand(names ...string) command {
	var cmd cmdVersion
	cmd.FlagSet = newCommand(&cmd, names...)
	setUsage(cmd.FlagSet, `(version|ver|v)

Print version information about the coin executable.`)
	return &cmd
}

func (cmd *cmdVersion) init() {}

func (cmd *cmdVersion) execute(w io.Writer) {
	fmt.Fprintln(w, "built on "+coin.Built)
	fmt.Fprintf(w, "built from %s@%s\n", coin.Branch, coin.Commit)
	fmt.Fprintln(w, "built with "+coin.GoVersion)
}
