package main

import (
	"testing"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/assert"
)

func Test_Pow(t *testing.T) {
	for i, tc := range []struct {
		b, e int
		r    int64
	}{
		{10, 1, 10},
		{10, 2, 100},
		{10, 5, 100000},
		{10, 11, 100000000000},
	} {
		assert.Equal(t, pow(tc.b, tc.e), tc.r, "%d not equal", i)
	}
}

func Test_AmtBetween(t *testing.T) {
	cad := &coin.Commodity{Id: "CAD", Decimals: 2}
	a := coin.MustParseAmount("10", cad)
	b := coin.MustParseAmount("1000000", cad)
	r := amtBetween(a, b)
	t.Error(r)
	assert.True(t, r.IsLessThan(b))
	assert.True(t, a.IsLessThan(r))
}
