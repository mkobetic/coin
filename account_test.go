package coin

import (
	"strings"
	"testing"

	"github.com/mkobetic/coin/assert"
)

func Test_AccountFromName(t *testing.T) {
	for _, f := range []struct {
		full, name, parent string
	}{
		{"Income:Salary", "Salary", "Income"},
		{"Expenses:Utilities:Water", "Water", "Expenses:Utilities"},
		{"Root", "Root", ""},
	} {
		a := accountFromName(f.full)
		assert.Equal(t, a.Name, f.name)
		assert.Equal(t, a.ParentName, f.parent)
	}
}

func Test_ParseAccount(t *testing.T) {
	r := strings.NewReader(`
account Assets:Investments:IVL:US
	note Investorline
	check commodity == USD
	ofx_bankid 200000100
	ofx_acctid 500766075509175102
`)
	p := NewParser(r)
	i, err := p.Next("")
	assert.NoError(t, err)
	a, ok := i.(*Account)
	assert.Equal(t, ok, true)
	assert.Equal(t, a.Name, "US")
	assert.Equal(t, a.FullName, "Assets:Investments:IVL:US")
	assert.Equal(t, a.ParentName, "Assets:Investments:IVL")
	assert.Equal(t, a.CommodityId, "USD")
	assert.Equal(t, a.Description, "Investorline")
	assert.Equal(t, a.OFXAcctId, "500766075509175102")
	assert.Equal(t, a.OFXBankId, "200000100")
}
