package coin

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"

	"github.com/mkobetic/coin/check"
)

type Rules interface {
	AccountFor(payee string) *Account
	Name() string
	Write(w io.Writer, max int) error
}

type RuleIndex struct {
	Accounts   map[string]*AccountRules // Maps ID to a set of rules for that account
	Sets       []*RuleSet
	SetsByName map[string]*RuleSet
}

func (rs *RuleIndex) AccountRulesFor(acctId string) *AccountRules {
	ars := rs.Accounts[acctId]
	if ars != nil {
		return ars
	}
	account := FindAccountOfxId(acctId)
	check.If(account != nil, "could not find account for Acct ID %s", acctId)
	return &AccountRules{Account: account}
}

func (rs *RuleIndex) Write(w io.Writer) error {
	for _, r := range rs.Sets {
		if _, err := fmt.Fprintf(w, "%s\n", r.Name()); err != nil {
			return err
		}
		if err := writeRules(r.Rules, w); err != nil {
			return err
		}
	}
	var accounts []*AccountRules
	ids := map[*AccountRules]string{}
	var max int
	for id, ars := range rs.Accounts {
		ids[ars] = id
		accounts = append(accounts, ars)
		if len(id) > max {
			max = len(id)
		}
	}
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Account.FullName < accounts[j].Account.FullName
	})
	for _, ars := range accounts {
		id := ids[ars]
		if _, err := fmt.Fprintf(w, "%-*s %s\n", max, id, ars.Account.FullName); err != nil {
			return err
		}
		if err := writeRules(ars.Rules, w); err != nil {
			return err
		}
	}
	return nil
}

func writeRules(rules []Rules, w io.Writer) error {
	var max int
	for _, r := range rules {
		if l := len(r.Name()); l > max {
			max = l
		}
	}
	for _, r := range rules {
		if err := r.Write(w, max); err != nil {
			return err
		}
	}
	return nil
}

// Rules to apply when importing transaction for given account
type AccountRules struct {
	Account *Account
	Rules   []Rules
}

func (ars *AccountRules) AccountFor(payee string) *Account {
	for _, r := range ars.Rules {
		acc := r.AccountFor(payee)
		if acc != nil {
			return acc
		}
	}
	return nil
}

type RuleSet struct {
	name  string
	Rules []Rules
}

func (rs *RuleSet) Name() string {
	return rs.name
}

func (rs *RuleSet) Write(w io.Writer, max int) error {
	_, err := fmt.Fprintf(w, "  @%-*s\n", max, rs.Name())
	return err
}

func (rs *RuleSet) AccountFor(payee string) *Account {
	for _, r := range rs.Rules {
		acc := r.AccountFor(payee)
		if acc != nil {
			return acc
		}
	}
	return nil
}

// If this rule matches the transaction description,
// use Account as the other side of the transaction
type Rule struct {
	Account *Account
	*regexp.Regexp
}

func (r *Rule) Name() string {
	return r.Account.FullName
}

func (r *Rule) Write(w io.Writer, max int) error {
	_, err := fmt.Fprintf(w, "  %-*s %s\n", max, r.Name(), r.String())
	return err
}

func (r *Rule) AccountFor(payee string) *Account {
	if r.MatchString(payee) {
		return r.Account
	}
	return nil
}

var patternRE = `([\w:$^\\-]+)`
var headerRE = regexp.MustCompile(`^(\w+)\s+` + patternRE + `|^(\w+)`)
var bodyRE = regexp.MustCompile(`^\s+` + patternRE + `(\s+(\S.*\S))?|^\s+@(\w+)`)

func ReadRules(r io.Reader) (*RuleIndex, error) {
	s := bufio.NewScanner(r)
	if !s.Scan() {
		return nil, s.Err()
	}
	return ScanRules(s.Bytes(), s)
}

func ScanRules(line []byte, s *bufio.Scanner) (*RuleIndex, error) {
	ri := &RuleIndex{
		Accounts:   make(map[string]*AccountRules),
		SetsByName: make(map[string]*RuleSet),
	}
	for {
		match := headerRE.FindSubmatch(line)
		if match != nil {
			var setRules func(rules []Rules)
			if len(match[1]) > 0 {
				ar := &AccountRules{Account: MustFindAccount(string(match[2]))}
				ri.Accounts[string(match[1])] = ar
				setRules = func(rules []Rules) { ar.Rules = rules }
			} else {
				rs := &RuleSet{name: string(match[3])}
				ri.Sets = append(ri.Sets, rs)
				ri.SetsByName[rs.Name()] = rs
				setRules = func(rules []Rules) { rs.Rules = rules }
			}
			var rules []Rules
			for {
				if !s.Scan() {
					setRules(rules)
					return ri, s.Err()
				}
				line = s.Bytes()
				match = bodyRE.FindSubmatch(line)
				if match == nil {
					break
				}
				if len(match[1]) > 0 {
					r := &Rule{
						Account: MustFindAccount(string(match[1])),
						Regexp:  regexp.MustCompile(string(match[3]))}
					rules = append(rules, r)
				} else {
					r := ri.SetsByName[string(match[4])]
					if r == nil {
						panic(fmt.Errorf("invalid rule set ref: %s", string(match[4])))
					}
					rules = append(rules, r)
				}
			}
			setRules(rules)
		} else {
			if !s.Scan() {
				return ri, s.Err()
			}
			line = bytes.TrimSpace(s.Bytes())
		}
	}
}

func stringify(m [][]byte) (o []string) {
	for _, b := range m {
		o = append(o, string(b))
	}
	return o
}
