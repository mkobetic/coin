package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/mkobetic/coin/assert"
)

func Test_Source(t *testing.T) {
	s := bufio.NewScanner(strings.NewReader(`mybank 1
  account "XXX"
  description 2
  date 1
  amount 3
  amount 4 -$1 VALUE = \s*([\d\.]+)\s*
  note 4
mybrokerage2 2
  account 0
  date 1
  amount 2
  note_symbol 4
  note_quantity 3
  note "${note_quantity} ${note_symbol}"
`))
	assert.True(t, s.Scan(), "Failed scanning first line: %s", s.Err())
	line := s.Bytes()
	src := ScanSource(line, s)
	assert.NotNil(t, src)
	assert.Equal(t, src.name, "mybank")
	assert.Equal(t, src.skip, 1)
	assert.Equal(t, len(src.fields), 5)
	get := func(name string, row ...string) string { return src.fields[name].Value(row, src.fields) }
	assert.Equal(t, get("amount", "AcctID", "12/11", "desc", "50.45", "note"), "50.45")
	assert.Equal(t, get("amount", "AcctID", "12/11", "desc", "", "note VALUE = 70.777"), "-70.777")
	assert.Equal(t, get("amount", "AcctID", "12/11", "desc", "", "note"), "")
	assert.Equal(t, get("date", "AcctID", "12/11", "desc", "50.45", "note"), "12/11")
	assert.Equal(t, get("account", "AcctID", "12/11", "desc", "50.45", "note"), "XXX")

	src = ScanSource(s.Bytes(), s)
	assert.NotNil(t, src)
	assert.Equal(t, src.name, "mybrokerage2")
	assert.Equal(t, src.skip, 2)
	assert.Equal(t, len(src.fields), 6)
	assert.Equal(t, get("note", "AcctID", "12/11", "desc", "50.45", "VBAL"), "50.45 VBAL")
}
