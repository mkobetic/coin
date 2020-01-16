package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mkobetic/coin"
	"github.com/pmezard/go-difflib/difflib"
)

func init() {
	(&cmdTest{}).newCommand("test", "t")
}

type cmdTest struct {
	*flag.FlagSet
}

func (_ *cmdTest) newCommand(names ...string) command {
	var cmd cmdTest
	cmd.FlagSet = newCommand(&cmd, names...)
	return &cmd
}

func (cmd *cmdTest) init() {
	coin.LoadFile(cmd.Arg(0))
	coin.ResolveAll()
}

func (cmd *cmdTest) execute(f io.Writer) {
	for _, t := range coin.Tests {
		var args []string
		scanner := bufio.NewScanner(bytes.NewReader(t.Cmd))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			args = append(args, scanner.Text())
		}
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "test item is missing command")
			return
		}
		command, found := aliases[args[0]]
		if !found {
			fmt.Fprintf(os.Stderr, "test command unknown: %s", args[0])
			return
		}
		command = command.newCommand(command.Name())
		if len(args) > 1 {
			command.Parse(args[1:])
		} else {
			command.Parse(nil)
		}
		var b bytes.Buffer
		command.execute(&b)
		if bytes.Equal(b.Bytes(), t.Result) {
			fmt.Fprintf(f, "OK %s %s\n", t.Location(), t.Cmd)
			continue
		}
		fmt.Fprintf(f, "FAIL %s %s\n", t.Location(), t.Cmd)
		difflib.WriteUnifiedDiff(f,
			difflib.UnifiedDiff{
				B:        difflib.SplitLines(b.String()),
				A:        difflib.SplitLines(string(t.Result)),
				FromFile: "expected",
				ToFile:   "actual",
				Context:  3,
			})
	}
}
