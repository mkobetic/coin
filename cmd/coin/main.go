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
	// constructor (used by test command)
	newCommand(names ...string) command

	// execution
	init()
	execute(io.Writer)

	// flag.FlagSet
	Name() string
	Parse(arguments []string) error
	Usage() // this is from flagsWithUsage
}

func newCommand(cmd command, names ...string) *flag.FlagSet {
	commands = append(commands, cmd)
	for _, n := range names {
		aliases[n] = cmd
	}
	return flag.NewFlagSet(names[0], flag.ExitOnError)
}

type flagsWithUsage struct{ *flag.FlagSet }

func (f flagsWithUsage) Usage() {
	f.FlagSet.Usage()
}

func setUsage(fs *flag.FlagSet, usage string) {
	fs.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, usage)
		fs.PrintDefaults()
	}
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
		if cmdArg != "-h" {
			fmt.Fprintf(os.Stderr, "Unknown command %s\n", cmdArg)
		}
		fmt.Fprintln(os.Stderr, "Usage: coin [cmd] ...")
		for _, c := range commands {
			fmt.Fprintf(os.Stderr, "\nCommand ")
			c.Usage()
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
