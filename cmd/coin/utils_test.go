package main

import (
	"testing"

	"github.com/mkobetic/coin/assert"
)

func Test_TrimWS(t *testing.T) {
	for _, tc := range []struct {
		in, out string
	}{
		{"   a  bb \n	ddd \n  ", "a bb\nddd\n"},
		{"xx[ 7   ]\n      !yyy", "xx[ 7 ]\n!yyy"},
	} {
		assert.Equal(t, trimWS(tc.in), tc.out)
	}
}
