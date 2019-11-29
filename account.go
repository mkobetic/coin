package coin

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"

	"github.com/mkobetic/coin/check/warn"
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

	OFXBankId string
	OFXAcctId string
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
	for _, line := range lines {
		_, err := io.WriteString(w, line)
		if err != nil {
			return err
		}
	}

	return nil
}

var AccountRE = `([A-Za-z][\w:/_\-]*\w)`
var accountHead = regexp.MustCompile(`account\s+` + AccountRE)
var accountBody = regexp.MustCompile(
	`(\s+(note)\s+(\S.+))|` +
		`(\s+(check)\s+(\S.+))|` +
		`(\s+(ofx_bankid)\s+(\d+))|` +
		`(\s+(ofx_acctid)\s+(\d+))`)
var checkCommodity = regexp.MustCompile(`commodity\s+==\s+` + CommodityRE)

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

func (p *Parser) parseAccount() (*Account, error) {
	matches := accountHead.FindSubmatch(p.Bytes())
	a := accountFromName(string(matches[1]))
	for p.Scan() {
		if len(bytes.TrimSpace(p.Bytes())) == 0 {
			return a, nil
		}
		matches = accountBody.FindSubmatch(p.Bytes())
		if matches == nil {
			return a, fmt.Errorf("Unrecognized account line: %s", p.Text())
		}
		switch {
		case bytes.Equal(matches[2], []byte("note")):
			a.Description = string(matches[3])
		case bytes.Equal(matches[5], []byte("check")):
			if matches = checkCommodity.FindSubmatch(matches[6]); matches != nil {
				a.CommodityId = string(matches[1])
			}
		case bytes.Equal(matches[8], []byte("ofx_bankid")):
			a.OFXBankId = string(matches[9])
		case bytes.Equal(matches[11], []byte("ofx_acctid")):
			a.OFXAcctId = string(matches[12])
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
