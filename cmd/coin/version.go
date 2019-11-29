package main

import (
	"fmt"
	"io"

	"github.com/mkobetic/coin"
)

var cmdVersion = newCommand(func() {}, version, "version", "ver", "v")

func version(w io.Writer) {
	fmt.Fprintln(w, "built on "+coin.Built)
	fmt.Fprintf(w, "built from %s@%s\n", coin.Branch, coin.Commit)
	fmt.Fprintln(w, "built with "+coin.GoVersion)
}
