package main

import (
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/aclindsa/ofxgo"
	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

var (
	dumpOFXIDs = flag.Bool("ids", false, "dump accounts with known ofx ids")
	dumpRules  = flag.Bool("rules", false, "dump the loaded account rules (useful for formatting)")
	bmoHack    = flag.Bool("bmo", false, "handle invalid qfx files from Bank of Montreal")
)

func main() {
	flag.Parse()

	coin.LoadFile(coin.CommoditiesFile)
	coin.LoadFile(coin.AccountsFile)
	coin.ResolveAccounts()

	if *dumpOFXIDs {
		coin.AccountsDo(func(a *coin.Account) {
			if a.OFXAcctId != "" {
				if a.OFXBankId != "" {
					fmt.Printf("%s/%s %s\n", a.OFXBankId, a.OFXAcctId, a.FullName)
				} else {
					fmt.Printf("%s %s\n", a.OFXAcctId, a.FullName)
				}
			}
		})
		return
	}

	var rules *coin.RuleIndex
	fn := filepath.Join(coin.DB, "ofx.rules")
	if _, err := os.Stat(fn); !os.IsNotExist(err) {
		file, err := os.Open(fn)
		check.NoError(err, "Failed to open %s", fn)
		defer file.Close()
		rules, err = coin.ReadRules(file)
		check.NoError(err, "Failed to parse %s", fn)
	}

	if *dumpRules {
		rules.Write(os.Stdout)
		return
	}

	var transactions []*coin.Transaction

	for _, fileName := range flag.Args() {
		file, err := os.Open(fileName)
		check.NoError(err, "Failed to open %s", fileName)
		defer file.Close()
		var r io.Reader = file
		if *bmoHack {
			r = newBMOReader(file)
		}
		batch, err := readTransactions(r, rules)
		check.NoError(err, "Cannot parse file %s", fileName)
		transactions = append(transactions, batch...)
	}
	// write transactions
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].Posted.Before(transactions[j].Posted)
	})

	for _, t := range transactions {
		t.Write(os.Stdout, false)
		fmt.Fprintln(os.Stdout)
	}
}

func readTransactions(r io.Reader, rules *coin.RuleIndex) (transactions []*coin.Transaction, err error) {
	responses, err := ofxgo.ParseResponse(r)
	if err != nil {
		return nil, err
	}

	// read bank transactions
	for _, resp := range responses.Bank {
		resp := resp.(*ofxgo.StatementResponse)
		rules := rules.AccountRulesFor(resp.BankAcctFrom.AcctID.String())
		last := len(resp.BankTranList.Transactions) - 1
		for i, t := range resp.BankTranList.Transactions {
			var balance *big.Rat
			if i == last {
				balance = &(resp.BalAmt.Rat)
			}
			transactions = append(transactions,
				newTransaction(rules,
					t.DtPosted.Time,
					t.Name.String(),
					t.TrnAmt.Rat,
					balance,
				))
		}
	}
	// read credit card transactions
	for _, resp := range responses.CreditCard {
		resp := resp.(*ofxgo.CCStatementResponse)
		rules := rules.AccountRulesFor(resp.CCAcctFrom.AcctID.String())
		for _, t := range resp.BankTranList.Transactions {
			transactions = append(transactions,
				newTransaction(rules,
					t.DtPosted.Time,
					t.Name.String(),
					t.TrnAmt.Rat,
					nil,
				))
		}
	}

	return transactions, nil
}

func newTransaction(ars *coin.AccountRules, date time.Time, payee string, amount big.Rat, balance *big.Rat) *coin.Transaction {
	from := ars.Account
	to := ars.AccountFor(payee)
	if to == nil {
		to = coin.Unbalanced
	}
	amt := coin.NewAmountFrac(amount.Num(), amount.Denom(), ars.Account.Commodity)
	var bal *coin.Amount
	if balance != nil {
		bal = coin.NewAmountFrac(balance.Num(), balance.Denom(), ars.Account.Commodity)
	}
	t := &coin.Transaction{
		Posted:      date,
		Description: payee}
	t.Post(from, to, amt, bal)
	return t
}
