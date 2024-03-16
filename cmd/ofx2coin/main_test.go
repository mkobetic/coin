package main

import (
	"io"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/assert"
)

func init() {
	r := strings.NewReader(`
commodity CAD
  format 1.00 CAD

commodity USD
  format 1.00 USD

account Assets:Bank:Checking
account Assets:Bank:Savings
account Expenses:Groceries
account Expenses:Auto
account Expenses:Auto:Gas
account Expenses:Miscellaneous
account Income:Salary
account Income:Interest
account Liabilities:Credit:MC
`)
	coin.Load(r, "")
	coin.ResolveAccounts()
}

func Test_Classification(t *testing.T) {
	r := strings.NewReader(sample)
	rules, err := coin.ReadRules(r)
	assert.NoError(t, err)
	date, _ := time.Parse("06/01/02", "18/10/20")
	for _, fix := range []struct {
		from  string
		payee string
		to    string
	}{
		{"479347938749398", "[TR] COSTCO WHOLESALE #9239", "Expenses:Groceries"},
		{"479347938749398", "JOE'S DINER", "Expenses:Miscellaneous"},
	} {
		tran := newTransaction(rules.Accounts[fix.from], date, fix.payee, *big.NewRat(-10000, 100), nil)
		if account := tran.Postings[0].Account.FullName; account != fix.to {
			t.Errorf("mismatched\nexp: %s\ngot: %s\n", fix.to, account)
		}
	}
	whatever := newTransaction(rules.Accounts["479347938749398"], date, "TO BE IGNORED WHEN WHATEVER", *big.NewRat(-10000, 100), nil)
	if whatever != nil {
		t.Error("should be nil")
	}
}

func Test_ReadTransactions(t *testing.T) {
	var r io.Reader
	r = strings.NewReader(sample)
	rules, err := coin.ReadRules(r)
	assert.NoError(t, err)
	mc := rules.SetsByName["common"]
	assert.NotNil(t, mc)
	assert.Equal(t, len(mc.Rules), 3)
	drop := mc.Rules[2].(*coin.Rule)
	if drop.Account != nil {
		t.Error("should be nil")
	}
	assert.True(t, drop.Match([]byte("TO BE IGNORED WHEN WHATEVER")))
	assert.Equal(t, len(drop.Notes), 1)
	assert.Equal(t, drop.Notes[0], "payee with WHATEVER in it will be ignored")

	r = strings.NewReader(txsSample)
	r = newBMOReader(r)
	txs, err := readTransactions(r, rules)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, len(txs), 10)
	for i, tx := range []string{
		`2019/01/04 [CK]NO.272
  Unbalanced             704.00 CAD
  Assets:Bank:Checking  -704.00 CAD
`,
		`2019/01/14 [CW]ROGERS CABLE TV
  Unbalanced             114.12 CAD
  Assets:Bank:Checking  -114.12 CAD
`,
		`2019/01/14 [CW]PRIMUS
  Unbalanced             54.18 CAD
  Assets:Bank:Checking  -54.18 CAD
`,
		`2019/01/14 [CW]ENBRIDGE
  Unbalanced             96.33 CAD
  Assets:Bank:Checking  -96.33 CAD
`,
		`2019/01/14 [CW]HYDRO
  Unbalanced             66.40 CAD
  Assets:Bank:Checking  -66.40 CAD
`,
		`2019/01/14 [CW]WATSWR
  Unbalanced             106.30 CAD
  Assets:Bank:Checking  -106.30 CAD
`,
		`2019/01/15 [DN]ACME PAY
  Assets:Bank:Checking   1211.04 CAD
  Income:Salary         -1211.04 CAD
`,
		`2019/01/15 [CW] TF
  Unbalanced             200.00 CAD
  Assets:Bank:Checking  -200.00 CAD
`,
		`2019/01/22 [DN]MEDICARE
  Assets:Bank:Checking   704.00 CAD
  Unbalanced            -704.00 CAD
`,
		`2019/01/28 [CW] TF
  Unbalanced             1059.51 CAD
  Assets:Bank:Checking  -1059.51 CAD = 11462.95 CAD
`,
	} {
		assert.Equal(t, txs[i].String(), tx)
	}
}

var sample = `
common
  Expenses:Groceries       FRESHCO|COSTCO WHOLESALE|FARM BOY|LOBLAWS
  Expenses:Auto:Gas        COSTCO GAS|PETROCAN|SHELL
  -- WHATEVER
  ; payee with WHATEVER in it will be ignored
389249328477983 Assets:Bank:Savings
  Income:Interest     Interest
392843029797099 Assets:Bank:Checking
    @common
	Income:Salary       ACME PAY 
479347938749398 Liabilities:Credit:MC
  Expenses:Auto            HUYNDAI|TOYOTA
  @common
  Expenses:Miscellaneous   
`

var txsSample = `
OFXHEADER:100
DATA:OFXSGML
VERSION:102
SECURITY:NONE
ENCODING:USASCII
CHARSET:1252
COMPRESSION:NONE
OLDFILEUID:NONE
NEWFILEUID:NONE

<OFX>
<SIGNONMSGSRSV1>
<SONRS>
<STATUS>
<CODE>0
<SEVERITY>INFO
<MESSAGE>OK
</STATUS>
<DTSERVER>20190127181602.926[-5:EDT]
<USERKEY>392843029797099
<LANGUAGE>ENG
<INTU.BID>00001
</SONRS>
</SIGNONMSGSRSV1>
<BANKMSGSETV1><BANKMSGSET>
<BANKMSGSRSV1>

<STMTTRNRS>
<TRNUID>392843029797099
<STATUS>
<CODE>0
<SEVERITY>INFO
<MESSAGE>OK
</STATUS>
<STMTRS>
<CURDEF>CAD
<BANKACCTFROM>
<BANKID>200000100
<ACCTID>392843029797099
<ACCTTYPE>CHECKING
</BANKACCTFROM>
<BANKTRANLIST>
<DTSTART>20190127000000.000[-5:EDT]
<DTEND>20190127000000.000[-5:EDT]

<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20190104000000.000[-5:EDT]
<TRNAMT>-704.0
<FITID>20190104192100

<CHECKNUM>272
<NAME>[CK]NO.272                                                                     

</STMTTRN>

<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20190114000000.000[-5:EDT]
<TRNAMT>-114.12
<FITID>20190113162900


<NAME>[CW]ROGERS CABLE TV                                                            

</STMTTRN>

<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20190114000000.000[-5:EDT]
<TRNAMT>-54.18
<FITID>20190113162901


<NAME>[CW]PRIMUS                                                                     

</STMTTRN>

<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20190114000000.000[-5:EDT]
<TRNAMT>-96.33
<FITID>20190113162902


<NAME>[CW]ENBRIDGE                                                                   

</STMTTRN>

<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20190114000000.000[-5:EDT]
<TRNAMT>-66.4
<FITID>20190113162903


<NAME>[CW]HYDRO                                                               

</STMTTRN>

<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20190114000000.000[-5:EDT]
<TRNAMT>-106.3
<FITID>20190113162904


<NAME>[CW]WATSWR                                                              

</STMTTRN>

<STMTTRN>
<TRNTYPE>CREDIT
<DTPOSTED>20190115000000.000[-5:EDT]
<TRNAMT>1211.04
<FITID>20190114214400


<NAME>[DN]ACME PAY                                                   

</STMTTRN>

<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20190115000000.000[-5:EDT]
<TRNAMT>-200.0
<FITID>20190115064000


<NAME>[CW] TF                                                         

</STMTTRN>

<STMTTRN>
<TRNTYPE>CREDIT
<DTPOSTED>20190122000000.000[-5:EDT]
<TRNAMT>704.0
<FITID>20190122172300


<NAME>[DN]MEDICARE                                               

</STMTTRN>

<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20190128000000.000[-5:EDT]
<TRNAMT>-1059.51
<FITID>20190127181200


<NAME>[CW] TF                                                    

</STMTTRN>

</BANKTRANLIST>
<LEDGERBAL>
<BALAMT>11462.95
<DTASOF>20190127181602.926[-5:EDT]
</LEDGERBAL>
<AVAILBAL>
<BALAMT>11462.95
<DTASOF>20190127181602.926[-5:EDT]
</AVAILBAL>
</STMTRS>
</STMTTRNRS>

</BANKMSGSRSV1>
<BANKMSGSET><BANKMSGSETV1>
</OFX>

`
