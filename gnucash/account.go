package gnucash

import (
	"regexp"
	"sort"
	"strings"

	"github.com/mkobetic/coin"
)

/*
Account = element gnc:account {
  attribute version { "2.0.0" },
  element act:name { text },
  element act:id { attribute type { "guid" }, GUID },
  element act:type { "NONE"
                   | "BANK"
                   | "CASH"
                   | "CREDIT"
                   | "ASSET"
                   | "LIABILITY"
                   | "STOCK"
                   | "MUTUAL"
                   | "CURRENCY"
                   | "INCOME"
                   | "EXPENSE"
                   | "EQUITY"
                   | "RECEIVABLE"
                   | "PAYABLE"
                   | "ROOT"
                   | "TRADING"
                   | "CHECKING"
                   | "SAVINGS"
                   | "MONEYMRKT"
                   | "CREDITLINE" },

  ( element act:commodity {
      element cmdty:space { text },
      element cmdty:id { text }
    },
    element act:commodity-scu { xsd:int },
    element act:non-standard-scu { empty }?
  )?,
  element act:code { text }?,
  element act:description { text }?,
  element act:slots { KvpSlot+ }?,
  element act:parent { attribute type { "guid" }, GUID }?,
  element act:lots { Lot+ }?
}
*/

type Account struct {
	Guid           string     `xml:"id"`
	Name           string     `xml:"name"`
	Type           string     `xml:"type"`
	CommoditySpace string     `xml:"commodity>space"`
	CommodityId    string     `xml:"commodity>id"`
	CommodityScu   int        `xml:"commodity-scu"`
	NonStandardScu int        `xml:"non-standard-scu,omitempty"`
	ParentGuid     string     `xml:"parent"`
	Code           string     `xml:"code,omitempty"`
	Description    string     `xml:"description,omitempty"`
	Slots          []*KvpSlot `xml:"slots>slot"`
}

func AccountFrom(a *Account) *coin.Account {
	var name string
	if a.Name == "Root Account" {
		name = "Root"
	} else {
		name = strings.Replace(a.Name, " ", "", -1)
	}
	return &coin.Account{
		Name:        name,
		Type:        a.Type,
		Code:        a.Code,
		Description: a.Description,
	}
}

var (
	AccountsByGuid     = map[string]*coin.Account{}
	AccountParentGuids = map[*coin.Account]string{}
)

func resolveAccounts(accounts []*Account) {
	for _, gca := range accounts {
		if gca.Guid == "" {
			panic("account without guid: " + gca.Name)
		}
		a := AccountFrom(gca)
		AccountsByGuid[gca.Guid] = a
		if gca.ParentGuid != "" {
			AccountParentGuids[a] = gca.ParentGuid
		} else {
			coin.Root = a
		}
		if gca.CommodityId != "" {
			a.Commodity = coin.MustFindCommodity(gca.CommodityId, "gnucash account")
		} else {
			a.Commodity = coin.DefaultCommodity()
		}
		if online_id := gca.slot("online_id"); online_id != "" {
			ids := gc_online_id.FindStringSubmatch(online_id)
			a.OFXBankId = ids[2]
			a.OFXAcctId = ids[3]
		}
	}
	// Link parents and children.
	for a, guid := range AccountParentGuids {
		a.Parent = mustFindAccount(guid)
		a.Parent.Children = append(a.Parent.Children, a)
	}

	// Sort children.
	for _, a := range AccountsByGuid {
		a.FullName = buildFullName(a)
		if a.Parent != nil {
			a.ParentName = a.Parent.FullName
		}
		coin.AccountsByName[a.FullName] = a
		sort.Slice(a.Children, func(i, j int) bool {
			return a.Children[i].Name < a.Children[j].Name
		})
	}
}

func isRoot(a *coin.Account) bool {
	return a.Type == "ROOT" || a.Parent == nil
}

func buildFullName(a *coin.Account) string {
	if a.FullName != "" {
		return a.FullName
	}
	if isRoot(a) || isRoot(a.Parent) {
		a.FullName = a.Name
	} else {
		a.FullName = buildFullName(a.Parent) + ":" + a.Name
	}
	return a.FullName
}

var gc_online_id = regexp.MustCompile(`((?P<bank>\d+)\s+)?(?P<acct>\d+)`)

func (a *Account) slot(key string) string {
	for _, s := range a.Slots {
		if s.Key == key {
			return s.Value.Value
		}
	}
	return ""
}

func mustFindAccount(guid string) *coin.Account {
	a := AccountsByGuid[guid]
	if a == nil {
		panic("unknown account guid: " + guid)
	}
	return a
}
