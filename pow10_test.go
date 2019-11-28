package coin

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/mkobetic/coin/assert"
)

func Test_Log10(t *testing.T) {
	for i, fix := range []struct {
		in  int64
		out int
	}{
		{-5, 0},
		{0, 0},
		{9, 0},
		{10, 1},
		{99, 1},
		{100, 2},
		{19999, 4},
		{1000000, 6},
		{math.MaxInt64, 18},
	} {
		l10 := log10(fix.in)
		assert.Equal(t, l10, fix.out, "%d. not equal", i)
	}
}

func Test_BigPow10(t *testing.T) {
	for i, fix := range []struct {
		in  string
		out int
	}{
		{"-555", 2},
		{"-5", 0},
		{"0", 0},
		{"9", 0},
		{"10", 1},
		{"99", 1},
		{"100", 2},
		{"19999", 4},
		{"1000000", 6},
		{strconv.FormatInt(math.MaxInt64, 10), 18},
		{strings.Repeat("11", 20), 39},
	} {
		bi, ok := new(big.Int).SetString(fix.in, 10)
		assert.Equal(t, ok, true)
		l10 := bigLog10(bi)
		assert.Equal(t, l10, fix.out, "%d. not equal", i)
	}
}
