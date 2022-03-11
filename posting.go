package coin

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Posting struct {
	Notes []string
	Tags  Tags

	Transaction     *Transaction
	Account         *Account
	Quantity        *Amount // posting amount
	Balance         *Amount // account balance as of this posting
	BalanceAsserted bool    // was balance explicitly asserted in the ledger

	accountName string
}

func (s *Posting) Write(w io.Writer, accountOffset, accountWidth, amountWidth int, ledger bool) error {
	notes := s.Notes
	commodity := s.Quantity.Commodity
	line := fmt.Sprintf("%*s%-*s  %*.*f %s",
		accountOffset, "",
		accountWidth, s.Account.FullName,
		amountWidth, commodity.Decimals, s.Quantity, commodity.SafeId(ledger))
	if s.BalanceAsserted {
		commodity = s.Balance.Commodity
		line += fmt.Sprintf(" = %.*f %s", commodity.Decimals, s.Balance, commodity.SafeId(ledger))
	}
	if len(notes) > 0 && len(notes[0])+len(line) < TRANSACTION_LINE_MAX-3 {
		line += " ; " + notes[0]
		notes = notes[1:]
	}
	err := writeStrings(w, nil, line, "\n")
	if err != nil {
		return err
	}
	for _, n := range notes {
		err := writeStrings(w, nil, "    ; ", n, "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Posting) String() string {
	var b strings.Builder
	s.Write(&b, 0, len(s.Account.FullName)+2, bigLog10(s.Quantity.Int)+3, false)
	return b.String()
}

func (s *Posting) IsEqual(s2 *Posting) bool {
	return s.Account == s2.Account &&
		s.Quantity.IsEqual(s2.Quantity)
}

func (p *Posting) MoveTo(a *Account) {
	if p.Account == a {
		return
	}
	p.Account.deletePosting(p)
	a.addPosting(p)
	p.Account = a
}

func (p *Posting) drop() {
	p.Account.deletePosting(p)
}

func (p *Posting) MarshalJSON() ([]byte, error) {
	var value = map[string]interface{}{
		"account":          p.Account.FullName,
		"quantity":         p.Quantity,
		"balance":          p.Balance,
		"balance_asserted": p.BalanceAsserted,
	}
	if len(p.Notes) > 0 {
		value["notes"] = p.Notes
	}
	if p.Tags != nil {
		value["tags"] = p.Tags
	}
	return json.MarshalIndent(value, "", "\t")
}
