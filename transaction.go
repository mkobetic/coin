package coin

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
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

var transactionRE = regexp.MustCompile(DateRE + `(\s+\((\w+)\))?(\s+(\S[^;]*))?(; ?(.*))?`)
var postingRE = regexp.MustCompile(`` +
	`\s+` + AccountRE + `(\s+` + AmountRE + `(\s+=\s+` + AmountRE + `)?)?|` +
	`\s+; ?(.*)`)

func (p *Parser) parseTransaction(fn string) (*Transaction, error) {
	match := transactionRE.FindSubmatch(p.Bytes())
	if match == nil {
		return nil, fmt.Errorf("Invalid transaction line: %s", p.Text())
	}
	t := &Transaction{
		Posted:      mustParseDate(match[1]),
		Code:        string(match[3]),
		Description: string(bytes.TrimRight(match[5], " \t")),
		Note:        string(bytes.TrimLeft(match[7], " \t")),
		line:        p.lineNr,
		file:        fn,
	}
	var notes []string
	var s *Posting
	for p.Scan() {
		match = postingRE.FindSubmatch(p.Bytes())
		if match == nil {
			break
		}
		if note := match[10]; len(note) > 0 {
			notes = append(notes, string(note))
			continue
		}
		var quantity *Amount
		var err error
		if len(match[3]) > 0 {
			c := MustFindCommodity(string(match[5]), t.Location())
			quantity, err = parseAmount(match[3], c)
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
			accountName: string(match[1]),
			Quantity:    quantity,
		}
		if balance := match[7]; len(balance) > 0 {
			c := MustFindCommodity(string(match[9]), t.Location())
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
	sFrom := &Posting{Account: from, Quantity: amount, Balance: balance}
	sTo := &Posting{Account: to, Quantity: amount.Negated()}
	if amount.Sign() < 0 {
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
