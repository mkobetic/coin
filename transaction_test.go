package coin

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mkobetic/coin/assert"
)

func Test_ParseTransaction(t *testing.T) {
	r := strings.NewReader(`
2008/04/02 COSTCO WHOLESALE
  Expenses:Groceries       37.92 CAD
  Liabilities:Credit:AMEX  -37.92 CAD

2008/04/03
  Assets:Investments:Martin:RRSP:Archive:TDB162  462.250 TDB162
  Assets:Investments:Martin:RRSP:CAD             -6000.00 CAD

2008/04/03 [PR]PC #7959 BELL'S
  Expenses:Groceries   125.66 CAD
  Assets:Bank:Checking  -125.66 CAD
`)
	p := NewParser(r)
	i, err := p.Next("")
	assert.NoError(t, err)
	tr, ok := i.(*Transaction)
	assert.Equal(t, ok, true)
	assert.Equal(t, tr.Description, "COSTCO WHOLESALE")
	assert.Equal(t, len(tr.Postings), 2)

	i, err = p.Next("")
	assert.NoError(t, err)
	tr, ok = i.(*Transaction)
	assert.Equal(t, ok, true)
	assert.Equal(t, tr.Description, "")
	assert.Equal(t, len(tr.Postings), 2)

	i, err = p.Next("")
	assert.NoError(t, err)
	tr, ok = i.(*Transaction)
	assert.Equal(t, ok, true)
	assert.Equal(t, tr.Description, "[PR]PC #7959 BELL'S")
	assert.Equal(t, len(tr.Postings), 2)

}

func Test_ParseTransactionNotes(t *testing.T) {
	r := strings.NewReader(`
; first example
; with a short note
2018/10/01 payee1 ; hello
  AA 10.00 CAD
    ; single-line split note
  BB -10.00 CAD; short note

; multi-line note
2018/10/02 payee2 ; hello
  ; hello world
  ;   and again
  ;and again
  AA 50.00 CAD
  BB 50.00 CAD; hi
    ; ho
  CC -100.00 CAD
	; multi-line
  ; split
    ;note
`)
	p := NewParser(r)
	i, err := p.Next("")
	assert.NoError(t, err)
	tr, ok := i.(*Transaction)
	assert.Equal(t, ok, true)
	assert.Equal(t, tr.Description, "payee1")
	assert.EqualStrings(t, tr.Notes, "hello")
	assert.Equal(t, len(tr.Postings), 2)
	assert.EqualStrings(t, tr.Postings[0].Notes, "single-line split note")
	assert.EqualStrings(t, tr.Postings[1].Notes, "short note")

	i, err = p.Next("")
	assert.NoError(t, err)
	tr, ok = i.(*Transaction)
	assert.Equal(t, ok, true)
	assert.Equal(t, tr.Description, "payee2")
	assert.EqualStrings(t, tr.Notes, "hello", "hello world", "  and again", "and again")
	assert.Equal(t, len(tr.Postings), 3)
	assert.EqualStrings(t, tr.Postings[1].Notes, "hi", "ho")
	assert.EqualStrings(t, tr.Postings[2].Notes, "multi-line", "split", "note")
}

func Test_ParseTransactionCode(t *testing.T) {
	r := strings.NewReader(`
2018/10/01 (code) payee1 ; hello
  AA 10.00 CAD
  BB -10.00 CAD
`)
	p := NewParser(r)
	i, err := p.Next("")
	assert.NoError(t, err)
	tr, ok := i.(*Transaction)
	assert.Equal(t, ok, true)
	assert.Equal(t, tr.Description, "payee1")
	assert.Equal(t, tr.Code, "code")
	assert.EqualStrings(t, tr.Notes, "hello")
	assert.Equal(t, len(tr.Postings), 2)
}

func Test_ParseTransactionBalance(t *testing.T) {
	r := strings.NewReader(`
2018/10/01 payee1
  AA 10.00 CAD
  BB -10.00 CAD = 50.00 CAD
`)
	p := NewParser(r)
	i, err := p.Next("")
	assert.NoError(t, err)
	tr, ok := i.(*Transaction)
	assert.Equal(t, ok, true)
	assert.Equal(t, tr.Description, "payee1")
	assert.Equal(t, len(tr.Postings), 2)
	assert.Equal(t, fmt.Sprintf("%a", tr.Postings[1].Balance), "50.00")
}

func Test_TransactionsByTimeDay(t *testing.T) {
	Year, Month, Day = 2000, 5, 7
	var transactions TransactionsByTime
	for _, days := range []int{0, 0, 0, 1, 4, 4, 4, 7, 13, 13, 20} {
		posted := MustParseDate(fmt.Sprintf("+%dd", days))
		transactions = append(transactions, &Transaction{Posted: posted})
	}
	check := func(t *testing.T, day time.Time, count int) {
		transactions := transactions.Day(day)
		assert.Equal(t, len(transactions), count, "%v length mismatch", day)
		for i, tr := range transactions {
			assert.Equal(t, tr.Posted, day, "%v transaction %d posted date mismatch", day, i)
		}
	}
	check(t, transactions[0].Posted.AddDate(0, 0, -10), 0)
	check(t, transactions[0].Posted.AddDate(0, 0, -1), 0)
	check(t, transactions[0].Posted, 3)
	check(t, transactions[3].Posted, 1)
	check(t, transactions[3].Posted.AddDate(0, 0, 1), 0)
	check(t, transactions[4].Posted, 3)
	check(t, transactions[7].Posted, 1)
	check(t, transactions[8].Posted, 2)
	check(t, transactions[10].Posted, 1)
	check(t, transactions[10].Posted.AddDate(0, 0, 1), 0)
	check(t, transactions[10].Posted.AddDate(0, 0, 10), 0)
}
