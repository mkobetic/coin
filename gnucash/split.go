package gnucash

import "github.com/mkobetic/coin"

/*
   <trn:split>
     <split:id type="guid">aeeeac5b036ac2ebb82dab8496761f71</split:id>
     <split:action>Buy</split:action>
     <split:reconciled-state>n</split:reconciled-state>
     <split:value>195300/100</split:value>
     <split:quantity>82612/1000</split:quantity>
     <split:account type="guid">d405662fc2e24e606a2c56f94f587737</split:account>
   </trn:split>

Split = element trn:split {
  element split:id { attribute type { "guid" }, GUID },
  element split:memo { text }?,
  element split:action { text }?,
  element split:reconciled-state { "y" | "n" | "c" | "f" | "v" },
  element split:reconcile-date { TimeSpec }?,
  element split:value { GncNumeric },
  element split:quantity { GncNumeric },
  element split:account { attribute type { "guid" }, GUID },
  element split:lot { attribute type { "guid" }, GUID }?,
  element split:slots { KvpSlot+ }?
}

*/

type Split struct {
	Guid             string `xml:"id"`
	AccountGuid      string `xml:"account"`
	Memo             string `xml:"memo,omitempty"`
	Action           string `xml:"action,omitempty"`
	ReconciledState  string `xml:"reconciled-state"`
	ReconciledStamp  string `xml:"reconcile-date>date,omitempty"`
	ValueFraction    string `xml:"value"`
	QuantityFraction string `xml:"quantity"`
}

func resolveSplits(splits []*Split, t *coin.Transaction) {
	for _, gs := range splits {
		s := &coin.Posting{Note: gs.Memo, Transaction: t}
		t.Postings = append(t.Postings, s)
		s.Account = mustFindAccount(gs.AccountGuid)
		s.Account.Postings = append(s.Account.Postings, s)
		s.Quantity = mustParseAmount(gs.QuantityFraction, s.Account.Commodity)
	}
}
