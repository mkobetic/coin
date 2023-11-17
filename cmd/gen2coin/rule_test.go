package main

import (
	"math/big"
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
	act := &coin.Account{Name: "A", Commodity: cad}
	a := big.NewInt(10)
	b := big.NewInt(1000000)
	r := amtBetween(int(a.Int64()), int(b.Int64()), act, nil).Int
	r = r.Div(r, big.NewInt(100)) // drop the 2 decimals
	assert.True(t, r.Cmp(b) == -1)
	assert.True(t, a.Cmp(r) == -1)
}
