package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/mkobetic/coin"
	"github.com/pmezard/go-difflib/difflib"
)

var (
	cmdTest *command
)

func init() {
	cmdTest = newCommand(test_load, test, "test", "t")
}

func test_load() {
	coin.LoadFile(cmdTest.Arg(0))
	coin.ResolveAll()
}

func test(f io.Writer) {
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
		if len(args) > 1 {
			command.Parse(args[1:])
		} else {
			command.Parse(nil)
		}
		var b bytes.Buffer
		command.execute(&b)
		if bytes.Equal(b.Bytes(), t.Result) {
			fmt.Fprintf(f, "%s ... OK\n", t.Cmd)
			continue
		}
		fmt.Fprintf(f, "%s ... FAIL\n", command.Name())
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
