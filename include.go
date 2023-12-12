package coin

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Include takes a path pattern that will be expanded into a list of files to include
// at the point when the include is encountered, i.e. the include files are loaded
// BEFORE the loading of the current files continues.
// The pattern is a glob pattern, e.g. "foo/*.coin" interpreted relative to the directory
// of the file containing the include statement.
// Absolute paths are interpreted relative to the directory from which the coin command was executed.
// Environment variables can be used in the pattern as well. They are expanded in the normal way (os.ExpandEnv).
type Include struct {
	Path string

	line uint
	file string
}

var includeHead = regexp.MustCompile(`include\s+(.+)`)

func (p *Parser) parseInclude(fn string) (*Include, error) {
	matches := includeHead.FindSubmatch(p.Bytes())
	i := &Include{Path: string(matches[1]), line: p.lineNr, file: fn}
	p.Scan()
	return i, nil
}

func (i *Include) Location() string {
	return fmt.Sprintf("%s:%d", i.file, i.line)
}

// Files returns list of file names matching the path to be included.
// Path resolution works as follows:
//   - leading / is prepended with . which is the directory from which the coin command was executed;
//   - otherwise the path is assumed to be relative to the directory of the file containing the include statement
//   - environment variable references are also expanded in normal way (os.ExpandEnv)
//   - if the path starts with an env var, it is left to be as it expands
//
// Finally the path is passed to filepath.Glob to expand wildcards.
func (i *Include) Files() ([]string, error) {
	var resolved string
	if i.Path[0] == '/' {
		resolved = "." + i.Path
	} else if i.Path[0] == '$' {
		resolved = i.Path
	} else {
		resolved = filepath.Join(filepath.Dir(i.file), i.Path)
	}
	resolved = os.ExpandEnv(resolved)
	return filepath.Glob(resolved)
}
