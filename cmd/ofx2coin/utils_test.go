package main

import (
	"testing"

	"github.com/mkobetic/coin/assert"
)

func Test_Trim(t *testing.T) {
	for _, tc := range []struct {
		in, out string
	}{
		{"   a  bb 	ddd   ", "a bb ddd"},
		{"xx[ 7   ]      !yyy", "xx[ 7 ] !yyy"},
	} {
		assert.Equal(t, trim(tc.in), tc.out)
	}
}
