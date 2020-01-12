package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	commands = []*command{}
	aliases  = map[string]*command{}
	verbose  bool
)

type command struct {
	*flag.FlagSet
	execute func(io.Writer)
	init    func()
}

func newCommand(init func(), cmd func(io.Writer), names ...string) *command {
	fs := flag.NewFlagSet(names[0], flag.ExitOnError)
	// can add common flags here
	fs.BoolVar(&verbose, "v", false, "output debugging info to stderr")
	c := &command{FlagSet: fs, execute: cmd, init: init}
	commands = append(commands, c)
	for _, n := range names {
		aliases[n] = c
	}
	return c
}

func debugf(format string, args ...interface{}) {
	if !verbose {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}
