package main

import (
	"flag"
	"io"
)

var (
	commands = []*command{}
	aliases  = map[string]*command{}
)

type command struct {
	*flag.FlagSet
	execute func(io.Writer)
	init    func()
}

func newCommand(init func(), cmd func(io.Writer), names ...string) *command {
	fs := flag.NewFlagSet(names[0], flag.ExitOnError)
	// can add common flags here
	c := &command{FlagSet: fs, execute: cmd, init: init}
	commands = append(commands, c)
	for _, n := range names {
		aliases[n] = c
	}
	return c
}
