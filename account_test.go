package coin

import (
	"math/big"
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
		assert.Equal(t, a.ParentName(), f.parent)
	}
}

func Test_ParseAccount(t *testing.T) {
	r := strings.NewReader(`
account Assets:Investments:IVL:US
	note Investorline
	commodity USD
	closed 2000/10/01
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
	assert.Equal(t, a.ParentName(), "Assets:Investments:IVL")
	assert.Equal(t, a.CommodityId, "USD")
	assert.Equal(t, a.Description, "Investorline")
	assert.Equal(t, a.OFXAcctId, "500766075509175102")
	assert.Equal(t, a.OFXBankId, "200000100")
	assert.True(t, a.IsClosed())
	assert.Equal(t, "2000/10/01", a.Closed.Format(DateFormat))
}

func Test_Postings(t *testing.T) {
	a := accountFromName("A")
	p1 := newPosting("2000/03", a)
	i, found := a.findPosting(p1)
	assert.False(t, found)
	assert.Equal(t, i, 0)
	a.addPosting(p1)
	assert.Equal(t, len(a.Postings), 1)
	assert.Equal(t, p1, a.Postings[0])
	p2 := newPosting("2000/07", a)
	i, found = a.findPosting(p2)
	assert.False(t, found)
	assert.Equal(t, i, 1)
	a.addPosting(p2)
	assert.Equal(t, len(a.Postings), 2)
	assert.Equal(t, p2, a.Postings[1])
	i, found = a.findPosting(p1)
	assert.True(t, found)
	assert.Equal(t, i, 0)
	p3 := newPosting("2000/03", a)
	i, found = a.findPosting(p3)
	assert.False(t, found)
	assert.Equal(t, i, 1)
	a.addPosting(p3)
	assert.Equal(t, len(a.Postings), 3)
	assert.Equal(t, a.Postings[1], p3)
	p4 := newPosting("2000/05", a)
	i, found = a.findPosting(p4)
	assert.False(t, found)
	assert.Equal(t, 2, i)
	a.addPosting(p4)
	assert.Equal(t, len(a.Postings), 4)
	a.deletePosting(p3)
	assert.Equal(t, len(a.Postings), 3)
	assert.Equal(t, cap(a.Postings), 4)
	assert.Equal(t, a.Postings[1], p4)
	p5 := a.addPosting(newPosting("2000/09", a))
	assert.Equal(t, len(a.Postings), 4)
	assert.Equal(t, a.Postings[3], p5)
	p6 := a.addPosting(newPosting("2000/01", a))
	assert.Equal(t, len(a.Postings), 5)
	assert.Equal(t, a.Postings[0], p6)
}

func newPosting(date string, a *Account) *Posting {
	d := MustParseDate(date)
	return &Posting{
		Account:     a,
		Quantity:    &Amount{big.NewInt(int64(d.Month())), cad},
		Transaction: &Transaction{Posted: d},
	}
}
