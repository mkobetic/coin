package coin

import (
	"strings"
	"testing"

	"github.com/mkobetic/coin/assert"
)

func Test_ParseCommodity(t *testing.T) {
	r := strings.NewReader(`
commodity NBC814
  note Altamira Precision Canadian Index Fund
  format 1.0000 NBC814

commodity BND
  note Vanguard Total Bond Market ETF
  format 1 BND
`)
	p := NewParser(r)
	i, err := p.Next()
	assert.NoError(t, err)
	c, ok := i.(*Commodity)
	assert.Equal(t, ok, true)
	assert.Equal(t, c.Id, "NBC814")
	assert.Equal(t, c.Name, "Altamira Precision Canadian Index Fund")
	assert.Equal(t, c.Decimals, 4)

	i, err = p.Next()
	assert.NoError(t, err)
	c, ok = i.(*Commodity)
	assert.Equal(t, ok, true)
	assert.Equal(t, c.Id, "BND")
	assert.Equal(t, c.Name, "Vanguard Total Bond Market ETF")
	assert.Equal(t, c.Decimals, 0)
}
