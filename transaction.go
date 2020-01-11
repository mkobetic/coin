package coin

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mkobetic/coin/rex"
)

type Transaction struct {
	Code        string
	Description string
	Note        string
	Postings    []*Posting

	Posted time.Time

	currencyId string
	line       uint
	file       string
}

var Transactions []*Transaction

func (t *Transaction) Write(w io.Writer, ledger bool) error {
	var notes []string
	if t.Note != "" {
		notes = strings.Split(t.Note, "\n")
	}
	line := t.Posted.Format(DateFormat) + " "
	if t.Code != "" {
		line += "(" + t.Code + ") "
	}
	line += t.Description
	if len(notes) == 1 && len(notes[0])+len(line) < 80 {
		line += " ; " + notes[0]
		notes = nil
	}
	_, err := io.WriteString(w, line+"\n")
	if err != nil {
		return err
	}
	for _, n := range notes {
		_, err := io.WriteString(w, "  ; "+n+"\n")
		if err != nil {
			return err
		}
	}
	maxn, maxa := 0, 0
	for _, s := range t.Postings {
		if l := len(s.Account.FullName); l > maxn {
			maxn = l
		}
		if l := s.Quantity.Width(s.Account.Commodity.Decimals); l > maxa {
			maxa = l
		}
	}
	for _, s := range t.Postings {
		err = s.Write(w, 2, maxn, maxa, ledger)
		if err != nil {
			return err
		}
	}
	return nil
}

var transactionREX = rex.MustCompile(`%s(\s+\((?P<code>\w+)\))?(\s+(?P<description>\S[^;]*))?(; ?(?P<note>.*))?`, DateREX)
var postingREX = rex.MustCompile(``+
	`\s+%s(\s+%s(\s+=\s+%s)?)?|`+
	`\s+; ?(?P<note>.*)`,
	AccountREX, AmountREX, AmountREX)

func (p *Parser) parseTransaction(fn string) (*Transaction, error) {
	match := transactionREX.Match(p.Bytes())
	if match == nil {
		return nil, fmt.Errorf("Invalid transaction line: %s", p.Text())
	}
	t := &Transaction{
		Posted:      mustParseDate(match, 0),
		Code:        match["code"],
		Description: strings.TrimRight(match["description"], " \t"),
		Note:        strings.TrimLeft(match["note"], " \t"),
		line:        p.lineNr,
		file:        fn,
	}
	var notes []string
	var s *Posting
	for p.Scan() {
		match = postingREX.Match(p.Bytes())
		if match == nil {
			break
		}
		if note := match["note"]; len(note) > 0 {
			notes = append(notes, string(note))
			continue
		}
		var quantity *Amount
		var err error
		if amt := match["amount1"]; len(amt) > 0 {
			c := MustFindCommodity(match["commodity1"], t.Location())
			quantity, err = parseAmount(amt, c)
			if err != nil {
				return nil, err
			}
		}
		if len(notes) > 0 {
			if s == nil {
				t.Note = strings.Join(notes, "\n")
			} else {
				s.Note = strings.Join(notes, "\n")
			}
			notes = nil
		}
		s = &Posting{
			Transaction: t,
			accountName: match["account"],
			Quantity:    quantity,
		}
		if balance := match["amount2"]; len(balance) > 0 {
			c := MustFindCommodity(match["commodity2"], t.Location())
			s.Balance, err = parseAmount(balance, c)
			if err != nil {
				return nil, err
			}
		}
		t.Postings = append(t.Postings, s)
	}
	if len(notes) > 0 {
		s.Note = strings.Join(notes, "\n")
	}
	return t, p.Err()
}

func (t *Transaction) String() string {
	var b strings.Builder
	t.Write(&b, false)
	return b.String()
}

func (t *Transaction) Location() string {
	return fmt.Sprintf("%s:%d", t.file, t.line)
}

func (t *Transaction) Post(
	from *Account,
	to *Account,
	amount *Amount,
	balance *Amount,
) {
	t.PostConversion(from, amount, balance, to, amount.Negated(), nil)
}

func (t *Transaction) PostConversion(
	from *Account,
	fromAmount *Amount,
	fromBalance *Amount,
	to *Account,
	toAmount *Amount,
	toBalance *Amount,
) {
	sFrom := &Posting{Account: from, Quantity: fromAmount, Balance: fromBalance}
	sTo := &Posting{Account: to, Quantity: toAmount, Balance: toBalance}
	if fromAmount.Sign() < 0 {
		t.Postings = append(t.Postings, sTo, sFrom)
	} else {
		t.Postings = append(t.Postings, sFrom, sTo)
	}
}

func (t *Transaction) Other(s *Posting) *Posting {
	for _, ss := range t.Postings {
		if ss != s {
			return ss
		}
	}
	return nil
}

func (t *Transaction) IsEqual(t2 *Transaction) bool {
	if !t.Posted.Equal(t2.Posted) {
		return false
	}
	if t.Code != t2.Code {
		return false
	}
	if len(t.Postings) != len(t2.Postings) {
		return false
	}
	for i, s := range t.Postings {
		if !t2.Postings[i].IsEqual(s) {
			return false
		}
	}
	return true
}
