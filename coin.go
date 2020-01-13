package coin

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mkobetic/coin/check"
)

const (
	// Default file names and extensions
	CoinExtension         = ".coin"
	AccountsFilename      = "accounts" + CoinExtension
	CommoditiesFilename   = "commodities" + CoinExtension
	PricesFilename        = "prices" + CoinExtension
	PricesExtension       = ".prices"
	TransactionsFilename  = "transactions" + CoinExtension
	TransactionsExtension = CoinExtension
)

var (
	DB                 = os.Getenv("COINDB")
	DefaultCommodityId = "CAD"

	AccountsFile     = filepath.Join(DB, AccountsFilename)
	CommoditiesFile  = filepath.Join(DB, CommoditiesFilename)
	PricesFile       = filepath.Join(DB, PricesFilename)
	TransactionsFile = filepath.Join(DB, TransactionsFilename)

	Tests []*Test

	Root           *Account
	Unbalanced     *Account
	AccountsByName = map[string]*Account{}

	// Build Parameters
	Built     string // time built in UTC
	Commit    string // source commit SHA
	Branch    string // source branch
	GoVersion string // Go version used to build
)

func DefaultCommodity() *Commodity {
	return MustFindCommodity(DefaultCommodityId, "default commodity")
}

func MustFindCommodity(id string, location string) *Commodity {
	if c := Commodities[id]; c != nil {
		return c
	}
	panic(fmt.Errorf("Can't find commodity %s\n\t%s\n", id, location))
}

func LoadPrices() {
	LoadFile(CommoditiesFile)
	if _, err := os.Stat(PricesFile); os.IsNotExist(err) {
		files, _ := filepath.Glob(filepath.Join(DB, "*.prices"))
		for _, f := range files {
			LoadFile(f)
		}
	} else {
		LoadFile(PricesFile)
	}
}

func LoadAll() {
	LoadPrices()
	LoadFile(AccountsFile)
	if _, err := os.Stat(TransactionsFile); os.IsNotExist(err) {
		files, _ := filepath.Glob(filepath.Join(DB, "*.coin"))
		for _, f := range files {
			if f != CommoditiesFile && f != AccountsFile {
				LoadFile(f)
			}
		}
	} else {
		LoadFile(TransactionsFile)
	}
	ResolveAll()
}

func LoadFile(filename string) {
	file, err := os.Open(filename)
	check.NoError(err, "Failed to open %s", filename)
	defer file.Close()
	Load(file, filename)
}

func Load(r io.Reader, fn string) {
	p := NewParser(r)
	for {
		i, err := p.Next(fn)
		check.NoError(err, "Parsing error")
		if i == nil {
			return
		}
		switch i := i.(type) {
		case *Commodity:
			Commodities[i.Id] = i
			if i.Symbol != "" {
				CommoditiesBySymbol[i.Symbol] = i
			}
		case *Account:
			if i.FullName == "" {
				panic(fmt.Errorf("INVALID %#v", i))
			}
			AccountsByName[i.FullName] = i
		case *Price:
			Prices = append(Prices, i)
		case *Transaction:
			Transactions = append(Transactions, i)
		case *Test:
			Tests = append(Tests, i)
		default:
			fmt.Fprintf(os.Stderr, "Unknown entity %T", i)
			os.Exit(1)

		}
	}
}

func ResolveAll() {
	ResolvePrices()
	ResolveAccounts()
	ResolveTransactions(true)
}

func ResolvePrices() {
	for _, p := range Prices {
		p.Commodity = MustFindCommodity(p.CommodityId, p.Location())
		p.Currency = MustFindCommodity(p.currencyId, p.Location())
		p.Commodity.AddPrice(p)
	}
	// Sort commodity prices.
	for _, c := range Commodities {
		for _, p := range c.Prices {
			sort.Slice(p, func(i, j int) bool {
				return p[i].Time.After(p[j].Time)
			})
		}
	}
	sort.Slice(Prices, func(i, j int) bool {
		return Prices[i].Time.Before(Prices[j].Time)
	})
}

func ResolveAccounts() {
	if Root == nil {
		Root = AccountsByName["Root"]
		if Root == nil {
			Root = accountFromName("Root")
			AccountsByName["Root"] = Root
		}
	}
	if Unbalanced == nil {
		Unbalanced = accountFromName("Unbalanced")
		AccountsByName["Unbalanced"] = Unbalanced
		// Root.adopt(Unbalanced)
	}
	// link parents with children, create parents if missing
	var known []*Account
	for _, a := range AccountsByName {
		known = append(known, a)
	}
LOOP:
	for _, a := range known {
		for {
			if a == Root {
				continue LOOP
			}
			if a.ParentName == "" {
				Root.adopt(a)
				continue LOOP
			}
			p := AccountsByName[a.ParentName]
			if p != nil {
				p.adopt(a)
				continue LOOP
			}
			p = accountFromName(a.ParentName)
			AccountsByName[p.FullName] = p
			p.adopt(a)
			a = p
		}
	}
	// sort children, link commodities
	for _, a := range AccountsByName {
		if a.CommodityId != "" {
			a.Commodity = MustFindCommodity(a.CommodityId, a.Location())
		} else {
			a.Commodity = DefaultCommodity()
		}
		sort.Slice(a.Children, func(i, j int) bool {
			return a.Children[i].Name < a.Children[j].Name
		})
	}

}

func ResolveTransactions(checkPostings bool) {
	for _, t := range Transactions {
		// t.Currency = MustFindCommodity(t.currencyId)
		var commodity *Commodity
		var commodities = map[*Commodity]bool{}
		for _, s := range t.Postings {
			s.Account = MustFindAccount(s.accountName)
			s.Account.Postings = append(s.Account.Postings, s)
			commodity = s.Account.Commodity
			commodities[commodity] = true
		}
		if len(commodities) > 1 {
			// Postings with different commodities, make sure amounts are set
			for _, s := range t.Postings {
				check.If(s.Quantity != nil, "Posting without quantity in mixed transaction: %s", t.Location())
			}
			continue
		}
		// All postings with the same commodity make sure transaction is balanced
		var empty *Posting
		var total = NewAmount(big.NewInt(0), commodity)
		for _, s := range t.Postings {
			if s.Quantity == nil {
				check.If(empty == nil, "Multiple postings without quantity: %s", t.Location())
				empty = s
			} else {
				total.AddIn(s.Quantity)
			}
		}
		if empty == nil {
			check.If(total.IsZero(), "Transaction is not balanced %f", total)
		} else {
			empty.Quantity = total.Negated()
		}
	}

	for _, a := range AccountsByName {
		sort.SliceStable(a.Postings, func(i, j int) bool {
			return a.Postings[i].Transaction.Posted.Before(a.Postings[j].Transaction.Posted)
		})
		if checkPostings {
			a.CheckPostings()
		}
	}
	sort.SliceStable(Transactions, func(i, j int) bool {
		return Transactions[i].Posted.Before(Transactions[j].Posted)
	})
}

func MustFindAccount(fullName string) *Account {
	if a := AccountsByName[fullName]; a != nil {
		return a
	}
	as := FindAccounts(fullName)
	if len(as) == 1 {
		return as[0]
	}
	if len(as) > 1 {
		msg := fmt.Sprintf("Found %d accounts matching %s", len(as), fullName)
		for _, a := range as {
			msg += "\n" + a.FullName
		}
		panic(msg)
	}
	panic(fmt.Errorf("Can't find account %s", fullName))
}

func FindAccountOfxId(acctId string) *Account {
	for _, a := range AccountsByName {
		if a.OFXAcctId == acctId {
			return a
		}
	}
	return nil
}

func ToRegex(pattern string) *regexp.Regexp {
	multiple := `[\w/_:-]*`
	single := `[\w/_-]*:[\w/_-]*`
	words := strings.Split(pattern, ":")
	rx := `(?i)` + words[0]
	if len(words) > 1 {
		wordLast := true
		for _, w := range words[1:] {
			if w == "" {
				if wordLast {
					rx += multiple
				}
				wordLast = false
			} else {
				if wordLast {
					rx += single
				}
				rx += w
				wordLast = true
			}
		}
	}
	return regexp.MustCompile(rx)
}

func FindAccounts(pattern string) (accounts []*Account) {
	rx := ToRegex(pattern)
	AccountsDo(func(a *Account) {
		if rx.MatchString(a.FullName) {
			accounts = append(accounts, a)
		}
	})
	return accounts
}

func CommoditiesDo(f func(c *Commodity)) {
	var names []string
	for n := range Commodities {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		f(Commodities[n])
	}
}

func AccountsDo(f func(c *Account)) {
	var names []string
	for n := range AccountsByName {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		f(AccountsByName[n])
	}
}
