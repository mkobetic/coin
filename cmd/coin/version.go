package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/mkobetic/coin"
)

func init() {
	(&cmdVersion{}).newCommand("version", "ver", "v")
}

type cmdVersion struct {
	*flag.FlagSet
}

func (*cmdVersion) newCommand(names ...string) command {
	var cmd cmdVersion
	cmd.FlagSet = newCommand(&cmd, names...)
	return &cmd
}

func (cmd *cmdVersion) init() {}

func (cmd *cmdVersion) execute(w io.Writer) {
	fmt.Fprintln(w, "built on "+coin.Built)
	fmt.Fprintf(w, "built from %s@%s\n", coin.Branch, coin.Commit)
	fmt.Fprintln(w, "built with "+coin.GoVersion)
}
