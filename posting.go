package coin

import (
	"fmt"
	"io"
	"strings"
)

type Posting struct {
	Note string

	Transaction *Transaction
	Account     *Account
	Quantity    *Amount // posting amount
	Balance     *Amount // account balance as of this posting
	Reconciled  bool    // was balance explicitly asserted in the ledger (only set after ResolveTransactions())

	accountName string
}

func (s *Posting) Write(w io.Writer, accountOffset, accountWidth, amountWidth int, ledger bool) error {
	commodity := s.Quantity.Commodity
	_, err := fmt.Fprintf(w, "%*s%-*s  %*.*f %s",
		accountOffset, "",
		accountWidth, s.Account.FullName,
		amountWidth, commodity.Decimals, s.Quantity, commodity.SafeId(ledger))
	if err != nil {
		return err
	}
	if s.Reconciled {
		if _, err = io.WriteString(w, " = "); err != nil {
			return err
		}
		if err = s.Balance.Write(w, ledger); err != nil {
			return err
		}
	}
	if _, err = io.WriteString(w, "\n"); err != nil {
		return err
	}
	if s.Note != "" {
		for _, n := range strings.Split(s.Note, "\n") {
			_, err := io.WriteString(w, "    ; "+n+"\n")
			if err != nil {
				return err
			}
		}

	}
	return err
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
