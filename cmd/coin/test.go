package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mkobetic/coin"
	"github.com/pmezard/go-difflib/difflib"
)

func init() {
	(&cmdTest{}).newCommand("test", "t")
}

type cmdTest struct {
	flagsWithUsage
	verbose bool
}

func (*cmdTest) newCommand(names ...string) command {
	var cmd cmdTest
	cmd.FlagSet = newCommand(&cmd, names...)
	setUsage(cmd.FlagSet, `(test|t)

Execute any test clauses found in the ledger (see tests/ directory).`)
	cmd.BoolVar(&cmd.verbose, "v", false, "print OK result for every test, not just for each file")
	return &cmd

}

func (cmd *cmdTest) init() {
	coin.LoadFile(cmd.Arg(0))
	coin.ResolveAll()
}

func (cmd *cmdTest) execute(f io.Writer) {
	if len(coin.Tests) == 0 {
		return
	}
	lastTestFile := file(coin.Tests[0])
	success := true
	for _, t := range coin.Tests {
		var args []string
		scanner := bufio.NewScanner(strings.NewReader(t.Cmd))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			args = append(args, scanner.Text())
		}
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "FAIL: test item is missing command %s\n", t.Location())
			return
		}
		// fmt.Println(args)
		// continue
		command, found := aliases[args[0]]
		if !found {
			fmt.Fprintf(os.Stderr, "FAIL: command unknown: %s %s\n", args[0], t.Location())
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
			testFile := file(t)
			if cmd.verbose {
				fmt.Fprintf(f, "OK %s %s\n", t.Location(), t.Cmd)
			} else if lastTestFile != testFile {
				fmt.Fprintf(f, "OK %s\n", lastTestFile)
			}
			lastTestFile = testFile
			continue
		}
		success = false
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
	if !cmd.verbose {
		result := "OK"
		if !success {
			result = "FAIL"
		}
		fmt.Fprintf(f, "%s %s\n", result, lastTestFile)
	}
}

func file(t *coin.Test) string {
	file, _, _ := strings.Cut(t.Location(), ":")
	return file
}
