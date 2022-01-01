package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
	"github.com/mkobetic/coin/rex"
)

var separator = []byte("---")

type Field struct {
	idx       *int     // field index for direct fields
	out       string   // field output template
	condField string   // names the field to match for conditional derived fields
	re        *rex.Exp // extraction rex for direct fields or conditional rex for derived fields
}

func (f *Field) Value(row []string, fields map[string]Fields) string {
	if f.idx == nil {
		return f.derivedField(row, fields)
	}
	s := row[*f.idx]
	if f.re == nil {
		return s
	}
	return f.directField(s)
}

func (f *Field) derivedField(row []string, fields map[string]Fields) string {
	if f.condField != "" && len(f.re.Match([]byte(fields[f.condField].Value(row, fields)))) == 0 {
		return ""
	}
	parts := strings.Split(f.out, "$")
	if len(parts) == 1 {
		return f.out
	}
	out := []string{parts[0]}
	for _, p := range parts[1:] {
		if p[0] == '{' {
			i := strings.IndexByte(p, '}')
			if i > 0 {
				f := p[1:i]
				out = append(out, fields[f].Value(row, fields)+p[i+1:])
				continue
			}
		}
		out = append(out, p)
	}
	return strings.Join(out, "")
}

func (f *Field) directField(s string) string {
	match := f.re.Match([]byte(s))
	if match == nil {
		return ""
	}
	parts := strings.Split(f.out, "$")
	if len(parts) == 1 {
		return f.out
	}
	out := []string{parts[0]}
	for _, p := range parts[1:] {
		if '0' <= p[0] && p[0] <= '9' {
			out = append(out, match[p[0:1]]+p[1:])
		} else {
			out = append(out, p)
		}
	}
	return strings.Join(out, "")
}

type Fields []*Field

func (fs Fields) Value(row []string, fields map[string]Fields) string {
	for _, f := range fs {
		if v := f.Value(row, fields); v != "" {
			return v
		}
	}
	return ""
}

type Source struct {
	name   string
	skip   int // number of header lines to skip
	fields map[string]Fields
}

var sourceREX = rex.MustCompile(`^(?P<source>\w+)(\s+(?P<skip>\d+))?\s*$`)
var derivedFieldRex = rex.MustCompile(`"(?P<code>.*)"(\s+(?P<condField>\w+)\s+(?P<condRex>.+))?`)
var directFieldRex = rex.MustCompile(`(?P<rowIdx>\d+)(\s+"(?P<out>.+)"\s+(?P<rex>.+))?`)
var fieldREX = rex.MustCompile(`^\s+(?P<field>\w+)\s+(%s|%s)$`, directFieldRex, derivedFieldRex)

func ScanSource(line []byte, s *bufio.Scanner) *Source {
	match := sourceREX.Match(line)
	if match == nil {
		return nil
	}
	skip, err := strconv.Atoi(match["skip"])
	check.NoError(err, "parsing header line count")
	src := &Source{name: match["source"], skip: skip, fields: make(map[string]Fields)}
	check.If(s.Scan(), "reading next source line: %s\n", s.Err())
	line = s.Bytes()
	for {
		match = fieldREX.Match(line)
		if match == nil {
			break
		}
		name := match["field"]
		var field Field
		if code := match["code"]; code != "" {
			field.out = code
			if field.condField = match["condField"]; field.condField != "" {
				field.re = rex.MustCompile(strings.TrimSpace(match["condRex"]))
			}
		} else {
			idx, err := strconv.Atoi(match["rowIdx"])
			check.NoError(err, "invalid field row index: %s\n", idx)
			field.idx = &idx
			field.out = strings.TrimSpace(match["out"])
			if ex := match["rex"]; ex != "" {
				field.re = rex.MustCompile(strings.TrimSpace(ex))
			}
		}
		src.fields[name] = append(src.fields[name], &field)
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
				if f.out != "" {
					fmt.Fprintf(w, "  %s %s\n", n, f.out)
				} else {
					fmt.Fprintf(w, "  %s %d\n", n, f.idx)
				}
			} else {
				fmt.Fprintf(w, "  %s %d %s %s\n", n, f.idx, f.out, f.re)
			}
		}
	}
}

func (s *Source) Value(field string, row []string) string {
	return s.fields[field].Value(row, s.fields)
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
			break
		}
		source := ScanSource(line, s)
		check.If(source != nil, "invalid source definition in rules file: %s", string(s.Bytes()))
		rules.sources[source.name] = source
		line = s.Bytes()
	}
	var err error
	rules.RuleIndex, err = coin.ScanRules(s.Bytes(), s)
	check.NoError(err, "Failed reading rules")
	return &rules
}
