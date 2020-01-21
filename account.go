package coin

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/mkobetic/coin/check"
	"github.com/mkobetic/coin/check/warn"
	"github.com/mkobetic/coin/rex"
)

type Account struct {
	Name        string
	FullName    string // name with all the ancestors
	Type        string
	Code        string
	Description string
	CommodityId string

	Commodity *Commodity
	Parent    *Account
	Children  []*Account
	Postings  []*Posting

	balance *Amount

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

func accountFromName(fullName string) *Account {
	_, name := parentAndName(fullName)
	return &Account{
		Name:        name,
		FullName:    fullName,
		CommodityId: DefaultCommodityId,
	}
}

func parentAndName(name string) (string, string) {
	i := strings.LastIndex(name, ":")
	if i < 0 {
		return "", name
	}
	return name[:i], name[i+1:]
}

func (a *Account) ParentName() string {
	parent, _ := parentAndName(a.FullName)
	return parent
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
	return fmt.Sprintf("%*a %-10s %s [%d]",
		a.Balance().Width(a.Commodity.Decimals),
		a.Balance(),
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

func (a *Account) Depth() int {
	if a.Parent == nil || a.Parent == Root {
		return 1
	}
	return a.Parent.Depth() + 1
}

func (a *Account) CheckPostings() {
	if len(a.Postings) == 0 {
		a.balance = NewZeroAmount(a.Commodity)
		return
	}
	a.balance = NewZeroAmount(a.Commodity)
	for _, s := range a.Postings {
		err := a.balance.AddIn(s.Quantity)
		check.NoError(err, "couldn't add %a %s to balance %a %s\n",
			s.Quantity, s.Quantity.Commodity, a.Balance(), a.Balance().Commodity)
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

func (a *Account) WithChildrenDo(f func(a *Account)) {
	f(a)
	for _, c := range a.Children {
		c.WithChildrenDo(f)
	}
}

func (a *Account) FirstWithChildrenDo(f func(a *Account)) {
	for _, c := range a.Children {
		c.FirstWithChildrenDo(f)
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

func ShortenAccountName(n string, size int) string {
	if len(n) <= size {
		return n
	}
	parts := strings.Split(n, ":")
	over := len(n) - size
	for i := 0; over > 0 && i < len(parts); i++ {
		l := len(parts[i])
		if l == 0 {
			continue
		}
		drop := min(over, l-1)
		parts[i] = parts[i][:l-drop]
		over -= drop
	}
	return strings.Join(parts, ":")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
