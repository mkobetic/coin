package coin

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mkobetic/coin/assert"
)

func init() {
	Commodities["TDB162"] = &Commodity{Id: "TDB162", Decimals: 4}
}

func Test_ParsePrice(t *testing.T) {
	r := strings.NewReader(`
P 1988/06/29 TDB162 9.69 CAD
P 1988/07/28 TDB162 9.58 CAD
P 1988/08/30 TDB162 9.45 CAD
`)
	p := NewParser(r)
	i, err := p.Next()
	assert.NoError(t, err)
	pr, ok := i.(*Price)
	assert.Equal(t, ok, true)
	assert.Equal(t, pr.CommodityId, "TDB162")
	assert.Equal(t, pr.currencyId, "CAD")
	assert.Equal(t, pr.Time.Format(DateFormat), "1988/06/29")
	assert.Equal(t, fmt.Sprintf("%a", pr.Value), "9.69")

	i, err = p.Next()
	assert.NoError(t, err)
	pr, ok = i.(*Price)
	assert.Equal(t, ok, true)
	assert.Equal(t, pr.CommodityId, "TDB162")
	assert.Equal(t, pr.currencyId, "CAD")
	assert.Equal(t, pr.Time.Format(DateFormat), "1988/07/28")
	assert.Equal(t, fmt.Sprintf("%a", pr.Value), "9.58")

	i, err = p.Next()
	assert.NoError(t, err)
	pr, ok = i.(*Price)
	assert.Equal(t, ok, true)
	assert.Equal(t, pr.CommodityId, "TDB162")
	assert.Equal(t, pr.currencyId, "CAD")
	assert.Equal(t, pr.Time.Format(DateFormat), "1988/08/30")
	assert.Equal(t, fmt.Sprintf("%a", pr.Value), "9.45")
}
