package main

import (
	"testing"

	"github.com/mkobetic/coin/assert"
)

func Test_TrimWS(t *testing.T) {
	for _, tc := range []struct {
		in, out []string
	}{
		{[]string{"   a  bb ", "	ddd ", "  "}, []string{"a bb", "ddd", ""}},
		{[]string{"xx[ 7   ]", "      !yyy"}, []string{"xx[ 7 ]", "!yyy"}},
	} {
		assert.EqualStrings(t, trimWS(tc.in...), tc.out...)
	}
}
