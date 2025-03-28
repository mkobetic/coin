package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
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

Execute any test clauses found in the ledger (see tests/ directory).
If test result is empty, updates the test file with computed result.`)
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
	var toUpdate []*coin.Test
	success := true
	for _, t := range coin.Tests {
		// assume the tests are sorted file by file
		// and in the order they are in the file
		testFile := file(t)
		startingNewFile := testFile != lastTestFile
		var oldTestFile string // this is set only when moving to new file
		if startingNewFile {
			oldTestFile = lastTestFile
		}
		lastTestFile = testFile
		if startingNewFile && len(toUpdate) > 0 {
			updateTestFile(toUpdate)
			toUpdate = nil
		}
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
		// If the test result is empty, update the test file
		// once we collect all the tests to be updated
		if len(t.Result) == 0 {
			t.Result = b.Bytes()
			toUpdate = append(toUpdate, t)
			continue
		}
		if bytes.Equal(b.Bytes(), t.Result) {
			if cmd.verbose {
				fmt.Fprintf(f, "OK %s %s\n", t.Location(), t.Cmd)
			} else if oldTestFile != "" {
				fmt.Fprintf(f, "OK %s\n", oldTestFile)
			}
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
	if len(toUpdate) > 0 {
		updateTestFile(toUpdate)
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

// assume ts are tests from the same file in the order in which they are in the file,
// and all test have empty result in the file,
// then write the results from the test commands into the file.
func updateTestFile(ts []*coin.Test) {
	fn := file(ts[0])
	tf, err := os.CreateTemp(path.Dir(fn), path.Base(fn))
	check.NoError(err, "creating temp file")
	// read file fn by lines up to line
	// write the lines into tf
	of, err := os.Open(fn)
	check.NoError(err, "opening old file")
	defer of.Close()

	scanner := bufio.NewScanner(of)
	writer := bufio.NewWriter(tf)
	line := 0
	for i, t := range ts {
		_, ln, _ := strings.Cut(t.Location(), ":")
		tLine, _ := strconv.Atoi(ln)

		for ; line < tLine && scanner.Scan(); line++ {
			_, err := writer.Write(scanner.Bytes())
			check.NoError(err, "writing prefix %d", i)
			check.NoError(writer.WriteByte('\n'), "writing prefix %d", i)
		}
		check.NoError(scanner.Err(), "writing prefix %d", i)
		_, err = writer.Write(t.Result)
		check.NoError(err, "writing result %d", i)
		fmt.Fprintf(os.Stderr, "UPDATED %s %s\n", t.Location(), t.Cmd)
	}
	// finish writing rest of the old file
	for scanner.Scan() {
		_, err := writer.Write(scanner.Bytes())
		check.NoError(err, "writing trailer")
		check.NoError(writer.WriteByte('\n'), "writing trailer")
	}
	check.NoError(scanner.Err(), "writing trailer")
	check.NoError(of.Close(), "closing old file")
	check.NoError(writer.Flush(), "flushing writer")
	check.NoError(tf.Close(), "closing temp file")
	check.NoError(os.Remove(fn), "deleting old file")
	check.NoError(os.Rename(tf.Name(), fn), "renaming temp file %s to %s", tf.Name(), fn)
}
