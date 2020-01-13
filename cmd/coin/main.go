package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
)

var (
	aliases  = map[string]command{}
	commands []command
)

type command interface {
	// constructor
	newCommand(...string) command

	// execution
	init()
	execute(io.Writer)

	// flag.FlagSet
	Parse(arguments []string) error
	PrintDefaults()
	Name() string
}

func newCommand(cmd command, names ...string) *flag.FlagSet {
	commands = append(commands, cmd)
	for _, n := range names {
		aliases[n] = cmd
	}
	return flag.NewFlagSet(names[0], flag.ExitOnError)
}

func main() {
	// resort commands alphabetically,
	// (needs to happen after they are all defined)
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})
	// default to balance command
	cmdArg := "balance"
	if len(os.Args) > 1 {
		cmdArg = os.Args[1]
	}
	cmd := aliases[cmdArg]
	if cmd == nil {
		fmt.Printf("Unknown command %s\n", cmdArg)
		for _, c := range commands {
			c.PrintDefaults()
		}
		os.Exit(1)
	}
	if len(os.Args) > 2 {
		cmd.Parse(os.Args[2:])
	} else {
		cmd.Parse(nil)
	}
	cmd.init()
	cmd.execute(os.Stdout)
}
