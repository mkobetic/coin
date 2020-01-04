package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/mkobetic/coin/assert"
)

func Test_Source(t *testing.T) {
	s := bufio.NewScanner(strings.NewReader(`mybank
  account 0
  description 2
  date 1
  amount 3
  amount 4 VALUE = \s*(?P<amount>[\d\.]+)\s*
  note 4
`))
	assert.True(t, s.Scan(), "Failed scanning first line: %s", s.Err())
	line := s.Bytes()
	src := ScanSource(line, s)
	assert.NotNil(t, src)
	assert.Equal(t, src.name, "mybank")
	assert.NotNil(t, src.fields)
	get := func(name string, row ...string) string { return src.fields[name].Value(name, row) }
	assert.Equal(t, get("amount", "AcctID", "12/11", "desc", "50.45", "note"), "50.45")
	assert.Equal(t, get("amount", "AcctID", "12/11", "desc", "", "note VALUE = 70.777"), "70.777")
	assert.Equal(t, get("amount", "AcctID", "12/11", "desc", "", "note"), "")
	assert.Equal(t, get("date", "AcctID", "12/11", "desc", "50.45", "note"), "12/11")
}
