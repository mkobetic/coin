package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
	"github.com/mkobetic/coin/rex"
)

var separator = []byte("---")

type Field struct {
	idx int
	re  *rex.Exp
}

func (f *Field) Value(name string, row []string) string {
	s := row[f.idx]
	if f.re == nil {
		return s
	}
	return f.re.Match([]byte(s))[name]
}

type Fields []*Field

func (fs Fields) Value(name string, row []string) string {
	for _, f := range fs {
		if v := f.Value(name, row); v != "" {
			return v
		}
	}
	return ""
}

type Source struct {
	name   string
	fields map[string]Fields
}

var sourceREX = rex.MustCompile(`^(?P<source>\w+)\s*$`)
var fieldREX = rex.MustCompile(`^\s+(?P<field>\w+)\s+(?P<rowIdx>\d+)(\s+(?P<rex>.*)\s*)?$`)

func ScanSource(line []byte, s *bufio.Scanner) *Source {
	match := sourceREX.Match(line)
	if match == nil {
		return nil
	}
	src := &Source{name: match["source"], fields: make(map[string]Fields)}
	check.If(s.Scan(), "reading next source line: %s\n", s.Err())
	line = s.Bytes()
	for {
		match = fieldREX.Match(line)
		if match == nil {
			break
		}
		name := match["field"]
		rowIdx, err := strconv.Atoi(match["rowIdx"])
		check.NoError(err, "invalid field row index: %s\n", match["rowIdx"])
		field := &Field{idx: rowIdx}
		if ex := match["rex"]; ex != "" {
			field.re = rex.MustCompile(ex)
		}
		src.fields[name] = append(src.fields[name], field)
		if !s.Scan() {
			break
		}
		line = s.Bytes()
	}
	return src
}

func (s *Source) Write(w io.Writer) {
	fmt.Fprintln(w, s.name)
	for _, n := range labels {
		fs := s.fields[n]
		for _, f := range fs {
			if f.re == nil {
				fmt.Fprintf(w, "  %s %d\n", n, f.idx)
			} else {
				fmt.Fprintf(w, "  %s %d %s\n", n, f.idx, f.re)
			}
		}
	}
}

type Rules struct {
	sources map[string]*Source
	*coin.RuleIndex
}

func (rules *Rules) Write(w io.Writer) {
	for _, src := range rules.sources {
		src.Write(w)
	}
	fmt.Fprintln(w, string(separator))
	rules.RuleIndex.Write(w)
}

var fieldsRE = regexp.MustCompile(`^(\w+)\s+([\d,]+)\s*$`)

func ReadRules(r io.Reader) *Rules {
	rules := Rules{sources: map[string]*Source{}}
	s := bufio.NewScanner(r)
	check.If(s.Scan(), "Failed scanning first line: %s", s.Err())
	line := s.Bytes()
	for {
		if bytes.Equal(line, separator) {
			if !s.Scan() { // nothing after separator
				return &rules
			}
			line = s.Bytes()
			break
		}
		source := ScanSource(line, s)
		check.If(source != nil, "invalid source definition in rules file")
		rules.sources[source.name] = source
		line = s.Bytes()
	}
	var err error
	rules.RuleIndex, err = coin.ScanRules(s.Bytes(), s)
	check.NoError(err, "Failed reading rules")
	return &rules
}
