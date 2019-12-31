package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/mkobetic/coin"
)

var separator = []byte("---")

type Rules struct {
	fields map[string][]int
	*coin.RuleIndex
}

func (rules *Rules) Write(w io.Writer) {
	for src, flds := range rules.fields {
		var fields []string
		for _, f := range flds {
			fields = append(fields, strconv.Itoa(f))
		}
		fmt.Fprintf(w, "%s %s\n", src, strings.Join(fields, ","))
	}
	fmt.Fprintf(w, "---\n")
	rules.RuleIndex.Write(w)
}

var fieldsRE = regexp.MustCompile(`^(\w+)\s+([\d,]+)\s*$`)

func ReadRules(r io.Reader) (*Rules, error) {
	rules := Rules{fields: map[string][]int{}}
	s := bufio.NewScanner(r)
	if !s.Scan() {
		return nil, s.Err()
	}
	line := s.Bytes()
	for {
		if bytes.Equal(line, separator) {
			if !s.Scan() {
				return nil, s.Err()
			}
			line = s.Bytes()
			break
		}
		match := fieldsRE.FindSubmatch(line)
		if match != nil {
			rules.fields[string(match[1])] = parseFields(string(match[2]))
		}
		if !s.Scan() {
			return nil, s.Err()
		}
		line = s.Bytes()
	}
	var err error
	rules.RuleIndex, err = coin.ScanRules(s.Bytes(), s)
	if err != nil {
		return nil, err
	}
	return &rules, nil
}
