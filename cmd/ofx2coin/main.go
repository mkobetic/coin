package main

import (
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sort"

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

	var rules *RuleIndex
	fn := filepath.Join(coin.DB, "ofx.rules")
	if _, err := os.Stat(fn); !os.IsNotExist(err) {
		file, err := os.Open(fn)
		check.NoError(err, "Failed to open %s", fn)
		defer file.Close()
		rules, err = ReadRules(file)
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

func readTransactions(r io.Reader, rules *RuleIndex) (transactions []*coin.Transaction, err error) {
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
				rules.Transaction(
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
				rules.Transaction(
					t.DtPosted.Time,
					t.Name.String(),
					t.TrnAmt.Rat,
					nil,
				))
		}
	}

	return transactions, nil
}
