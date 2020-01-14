package coin

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"strings"
	"unicode"

	"github.com/mkobetic/coin/check/warn"
	"github.com/mkobetic/coin/rex"
)

type Account struct {
	Name        string
	FullName    string // name with all the ancestors
	ParentName  string // full name of the parent account
	Type        string
	Code        string
	Description string
	CommodityId string

	Commodity *Commodity
	Parent    *Account
	Children  []*Account
	Postings  []*Posting

	balance           *Amount
	cumulativeBalance *Amount

	line uint
	file string

	OFXBankId string
	OFXAcctId string

	CSVAcctId string
}

/*
account Expenses:Food
	note This account is all about the chicken!
	alias food
	payee ^(KFC|Popeyes)$
	check commodity == "$"
	assert commodity == "$"
	eval print("Hello!")
	default
*/
func (a *Account) Write(w io.Writer, ledger bool) error {
	lines := []string{"account ", a.FullName, "\n"}
	if a.Description != "" {
		lines = append(lines, "  note ", a.Description, "\n")
	}
	lines = append(lines, `  check commodity == `, a.Commodity.SafeId(ledger), "\n")
	if a.OFXBankId != "" && !ledger {
		lines = append(lines, `  ofx_bankid `, a.OFXBankId, "\n")
	}
	if a.OFXAcctId != "" && !ledger {
		lines = append(lines, `  ofx_acctid `, a.OFXAcctId, "\n")
	}
	if a.CSVAcctId != "" && !ledger {
		lines = append(lines, `  csv_acctid `, a.CSVAcctId, "\n")
	}
	for _, line := range lines {
		_, err := io.WriteString(w, line)
		if err != nil {
			return err
		}
	}

	return nil
}

var accountNameREX = rex.MustCompile(`([A-Za-z][\w/_\-]*)`)
var AccountREX = rex.MustCompile(`(?P<account>%s(:%s)*)`, accountNameREX, accountNameREX)
var accountHeadREX = rex.MustCompile(`account\s+%s`, AccountREX)
var accountBodyREX = rex.MustCompile(``+
	`(\s+note\s+(?P<note>\S.+))|`+
	`(\s+check\s+commodity\s+==\s+%s|`+
	`(\s+ofx_bankid\s+(?P<ofx_bankid>\d+))|`+
	`(\s+ofx_acctid\s+(?P<ofx_acctid>\d+))|`+
	`(\s+csv_acctid\s+(?P<csv_acctid>\w+)))`,
	CommodityREX)

func accountFromName(name string) *Account {
	i := strings.LastIndex(name, ":")
	parent := ""
	if i > 0 {
		parent = name[:i]
	}
	return &Account{
		Name:        name[i+1:],
		FullName:    name,
		ParentName:  parent,
		CommodityId: DefaultCommodityId,
	}
}

func (p *Parser) parseAccount(fn string) (*Account, error) {
	match := accountHeadREX.Match(p.Bytes())
	a := accountFromName(match["account"])
	a.line = p.lineNr
	a.file = fn
	for p.Scan() {
		line := p.Bytes()
		if len(bytes.TrimSpace(line)) == 0 || !unicode.IsSpace(rune(line[0])) {
			return a, nil
		}
		match = accountBodyREX.Match(line)
		if match == nil {
			return a, fmt.Errorf("Unrecognized account line: %s", p.Text())
		}
		if n := match["note"]; n != "" {
			a.Description = n
		} else if c := match["commodity"]; c != "" {
			a.CommodityId = c
		} else if i := match["ofx_bankid"]; i != "" {
			a.OFXBankId = i
		} else if i := match["ofx_acctid"]; i != "" {
			a.OFXAcctId = i
		} else if i := match["csv_acctid"]; i != "" {
			a.CSVAcctId = i
		}
	}
	return a, p.Err()
}

func (a *Account) String() string {
	cum, err := a.CumulativeBalance()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%15a %15a %-10s %s [%d]",
		a.Balance(),
		cum,
		a.Commodity.Id,
		a.FullName,
		len(a.Postings))
}

func (a *Account) Location() string {
	return fmt.Sprintf("%s:%d", a.file, a.line)
}

func (a *Account) Balance() *Amount {
	return a.balance
}

func (a *Account) CheckPostings() {
	if len(a.Postings) == 0 {
		a.balance = NewAmount(big.NewInt(0), a.Commodity)
		return
	}
	a.balance = NewAmount(big.NewInt(0), a.Commodity)
	for _, s := range a.Postings {
		a.balance.AddIn(s.Quantity)
		if s.Balance != nil {
			warn.If(!a.balance.IsEqual(s.Balance),
				"%s: %s balance is %a, should be %a\n",
				a.FullName,
				s.Transaction.Posted.Format(DateFormat),
				a.balance,
				s.Balance,
			)
		}
	}
}

func (a *Account) CumulativeBalance() (*Amount, error) {
	if a.cumulativeBalance != nil {
		return a.cumulativeBalance, nil
	}
	total := a.Balance().Copy()
	for _, c := range a.Children {
		val, err := c.CumulativeBalance()
		if err != nil {
			return nil, err
		}
		val, err = a.Commodity.Convert(val, c.Commodity)
		if err != nil {
			return nil, err
		}
		total.AddIn(val)
	}
	return total, nil
}

func (a *Account) WithChildrenDo(f func(a *Account)) {
	f(a)
	for _, c := range a.Children {
		c.WithChildrenDo(f)
	}
}

func (a *Account) FirstWithChildrenDo(f func(a *Account)) {
	for _, c := range a.Children {
		c.WithChildrenDo(f)
	}
	f(a)
}

func (a *Account) adopt(c *Account) {
	isChild := false
	c.WithChildrenDo(func(d *Account) {
		isChild = isChild || (d == a)
	})
	if isChild {
		fmt.Printf("%#v\n", c)
		panic(fmt.Errorf("%s is child of %s", a.FullName, c.FullName))
	}
	a.WithChildrenDo(func(d *Account) {
		isChild = isChild || (d == c)
	})
	if isChild {
		fmt.Printf("%#v\n", c)
		panic(fmt.Errorf("%s is already a child of %s", c.FullName, a.FullName))
	}
	c.Parent = a
	a.Children = append(a.Children, c)
}
