package gnucash

import (
	"sort"
	"strings"

	"github.com/mkobetic/coin"
)

/*
<gnc:transaction version="2.0.0">
  <trn:id type="guid">1819adbc328c91036af266ace42d79b7</trn:id>
  <trn:currency>
    <cmdty:space>ISO4217</cmdty:space>
    <cmdty:id>CAD</cmdty:id>
  </trn:currency>
  <trn:date-posted>
    <ts:date>2007-12-12 05:59:00 -0500</ts:date>
  </trn:date-posted>
  <trn:date-entered>
    <ts:date>2010-02-14 23:50:26 -0500</ts:date>
  </trn:date-entered>
  <trn:description>Investment</trn:description>
  <trn:splits>
    <trn:split>
      <split:id type="guid">aeeeac5b036ac2ebb82dab8496761f71</split:id>
      <split:action>Buy</split:action>
      <split:reconciled-state>n</split:reconciled-state>
      <split:value>195300/100</split:value>
      <split:quantity>82612/1000</split:quantity>
      <split:account type="guid">d405662fc2e24e606a2c56f94f587737</split:account>
	</trn:split>
	...
  </trn:splits>
</gnc:transaction>

Transaction = element gnc:transaction {
  attribute version { "2.0.0" },
  element trn:id { attribute type { "guid" }, GUID },
  element trn:currency {
    element cmdty:space { text },
    element cmdty:id { text }
  },
  element trn:num { text }?,
  element trn:date-posted { TimeSpec },
  element trn:date-entered { TimeSpec },
  element trn:description { text }?,
  element trn:slots { KvpSlot+ }?,
  element trn:splits { Split+ }
}
*/

type Transaction struct {
	Guid          string   `xml:"id"`
	CurrencySpace string   `xml:"currency>space"`
	CurrencyId    string   `xml:"currency>id"`
	Num           string   `xml:"num,omitempty"`
	PostedStamp   string   `xml:"date-posted>date"`
	EnteredStamp  string   `xml:"date-entered>date"`
	Description   string   `xml:"description,omitempty"`
	Splits        []*Split `xml:"splits>split"`
}

func resolveTransactions(transactions []*Transaction) {
	for _, gt := range transactions {
		ds := strings.SplitN(gt.Description, " - ", 2)
		t := &coin.Transaction{
			Code:        gt.Num,
			Description: ds[0],
		}
		if len(ds) == 2 {
			t.Note = ds[1]
		}
		t.Posted = mustParseTimeStamp(gt.PostedStamp)
		resolveSplits(gt.Splits, t)
		coin.Transactions = append(coin.Transactions, t)
	}
	for _, a := range AccountsByGuid {
		sort.Slice(a.Postings, func(i, j int) bool {
			return a.Postings[i].Transaction.Posted.Before(a.Postings[j].Transaction.Posted)
		})
		a.CheckPostings()
	}
	sort.Slice(coin.Transactions, func(i, j int) bool {
		return coin.Transactions[i].Posted.Before(coin.Transactions[j].Posted)
	})
}
