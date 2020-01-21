package coin

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/mkobetic/coin/assert"
)

var cad *Commodity

func init() {
	Commodities["CAD"] = &Commodity{Id: "CAD", Decimals: 2}
	cad = DefaultCommodity()
}

func Test_ParseAmount(t *testing.T) {
	for i, fix := range []struct {
		in, out string
	}{
		{"10", "10.00"},
		{"-50.01", "-50.01"},
		{"0.011", "0.01"},
		{"-100.00", "-100.00"},
	} {
		amt, err := parseAmount(fix.in, cad)
		assert.NoError(t, err)
		res := fmt.Sprintf("%a", amt)
		assert.Equal(t, res, fix.out, "%d. not equal", i)
	}
}

func Test_AmountWidth(t *testing.T) {
	for i, fix := range []struct {
		amt             string
		decimals, width int
	}{
		{"0.09", 2, 4},
		{"-0.09", 2, 5},
	} {
		amt, err := parseAmount(fix.amt, cad)
		assert.NoError(t, err)
		width := amt.Width(fix.decimals)
		assert.Equal(t, width, fix.width, "%d. not equal", i)
	}
}

func Test_FormatWidthAmount(t *testing.T) {
	for i, fix := range []struct {
		amt             string
		width, decimals int
		out             string
	}{
		{"10", 5, 0, "   10"},
		{"-50.01", 10, 2, "    -50.01"},
		{"0.001", 2, 3, "0.000"},
	} {
		amt, err := parseAmount(fix.amt, cad)
		assert.NoError(t, err)
		res := fmt.Sprintf("%*.*f", fix.width, fix.decimals, amt)
		assert.Equal(t, res, fix.out, "%d. not equal", i)
	}
}

func Test_AmountAddIn(t *testing.T) {
	for i, fix := range []struct {
		a, b, c string
	}{
		{"100.0", "500.00", "600.00"},
		{"100.00", "-50.0", "50.00"},
		{"-100.00", "50", "-50.00"},
	} {
		a := MustParseAmount(fix.a, cad)
		b := MustParseAmount(fix.b, cad)
		err := a.AddIn(b)
		assert.NoError(t, err)
		c := fmt.Sprintf("%a", a)
		assert.Equal(t, c, fix.c, "%d. not equal", i)
	}
}

func Test_AmountTimes(t *testing.T) {
	for i, fix := range []struct {
		a, b, c string
	}{
		{"100.0", "50.00", "5000.00"},
		{"100.00", "-50.0", "-5000.00"},
		{"-100.00", "50", "-5000.00"},
	} {
		a := MustParseAmount(fix.a, cad)
		b := MustParseAmount(fix.b, cad)
		c := a.Times(b)
		cc := fmt.Sprintf("%a", c)
		assert.Equal(t, cc, fix.c, "%d. not equal", i)
	}
}

func Test_NewAmountFracDec(t *testing.T) {
	for i, fix := range []struct {
		a, b, c string
	}{
		{"-704", "1", "-704.00"},
		{"-2853", "25", "-114.12"},
		{"-2709", "50", "-54.18"},
		{"-9633", "100", "-96.33"},
		{"-332", "5", "-66.40"},
		{"-1063", "10", "-106.30"},
		{"105276", "25", "4211.04"},
		{"200", "1", "200.00"},
	} {
		a := mustParseBigInt(fix.a)
		b := mustParseBigInt(fix.b)
		c := NewAmountFrac(a, b, cad)
		cc := fmt.Sprintf("%a", c)
		assert.Equal(t, cc, fix.c, "%d. not equal", i)
	}
}

func mustParseBigInt(s string) *big.Int {
	i := new(big.Int)
	i.SetString(s, 10)
	return i
}
