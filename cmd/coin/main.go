package main

import (
	"fmt"
	"os"
	"sort"
)

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
	command := aliases[cmdArg]
	if command == nil {
		fmt.Printf("Unknown command %s\n", cmdArg)
		for _, c := range commands {
			c.Usage()
		}
		os.Exit(1)
	}
	if len(os.Args) > 2 {
		command.Parse(os.Args[2:])
	} else {
		command.Parse(nil)
	}
	command.init()
	command.execute(os.Stdout)
}
