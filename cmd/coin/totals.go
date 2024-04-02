package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

// total represents an amount total for a time period.
type total struct {
	time.Time
	*coin.Amount
}

// totals aggregates time series amounts for accounts.
// It must be initialized with a reducer before use.
type totals struct {
	// this must be set before using
	*reducer
	// these are internal
	all     []*total
	current *total
}

// add amount at time t, t must not be before ts.current.Time,
// i.e. items being added must be sorted by time.
func (ts *totals) add(t time.Time, a *coin.Amount) {
	period := ts.reduce(t)
	if ts.current != nil && ts.current.Equal(period) {
		err := ts.current.AddIn(a)
		check.NoError(err, "cannot add %a to totals", a)
		return
	}
	amt := a.Copy()
	ts.current = &total{Time: period, Amount: amt}
	ts.all = append(ts.all, ts.current)
}

func (ts *totals) addTotals(ts2 ...*total) {
	for _, t := range ts2 {
		ts.add(t.Time, t.Amount)
	}
}

func (ts *totals) addTimes(ts2 ...*total) {
	for _, t := range ts2 {
		ts.add(t.Time, coin.NewZeroAmount(t.Commodity))
	}
}

// merge two sorted totals
func (ts *totals) merge(ts2 *totals) {
	all1, all2 := ts.all, ts2.all
	ts.all, ts.current = nil, nil
	for {
		if len(all1) == 0 {
			ts.addTotals(all2...)
			return
		}
		if len(all2) == 0 {
			ts.addTotals(all1...)
			return
		}
		if all1[0].After(all2[0].Time) {
			ts.addTotals(all2[0])
			all2 = all2[1:]
		} else {
			ts.addTotals(all1[0])
			all1 = all1[1:]
		}
	}
}

// mergeTime backfills ts with missing times from ts2
func (ts *totals) mergeTime(ts2 *totals) {
	all1, all2 := ts.all, ts2.all
	ts.all, ts.current = nil, nil
	for {
		if len(all1) == 0 {
			ts.addTimes(all2...)
			return
		}
		if len(all2) == 0 {
			ts.addTotals(all1...)
			return
		}
		if all1[0].After(all2[0].Time) {
			ts.addTimes(all2[0])
			all2 = all2[1:]
		} else {
			ts.addTotals(all1[0])
			all1 = all1[1:]
		}
	}
}

// maxWidth returns the largest amount width needed for printing
func (ts *totals) maxWidth() int {
	if ts == nil || ts.current == nil {
		return 0
	}
	var max int
	for _, t := range ts.all {
		if w := t.Width(ts.current.Commodity.Decimals); w > max {
			max = w
		}

	}
	return max
}

// maxMagnitued returns the amount with largest absolute value
func (ts *totals) maxMagnitude() *coin.Amount {
	if len(ts.all) == 0 {
		return nil
	}
	max := ts.all[0].Amount
	if len(ts.all) == 1 {
		return max
	}
	for _, t := range ts.all[1:] {
		if t.IsBigger(max) {
			max = t.Amount
		}
	}
	return max
}

// cumMagnitude returns the sum of the totals
func (ts *totals) cumMagnitude() *coin.Amount {
	if len(ts.all) == 0 {
		return nil
	}
	total := coin.NewZeroAmount(ts.all[0].Commodity)
	for _, t := range ts.all {
		err := total.AddIn(t.Amount)
		check.NoError(err, "computing cumulative magnitude")
	}
	return total
}

// cumulative converts ts to a cumulative totals sequence
func (ts *totals) makeCumulative() {
	if ts.current == nil {
		return
	}
	all := ts.all
	cum := coin.NewZeroAmount(ts.current.Commodity)
	ts.all, ts.current = nil, nil
	for _, t := range all {
		err := cum.AddIn(t.Amount)
		check.NoError(err, "converting totals to cumulative")
		ts.add(t.Time, cum)
	}
}

func (ts *totals) validate(acc string) {
	check.If(ts.current != nil, "current nil for %s", acc)
	for _, t := range ts.all {
		check.If(t.Amount != nil, "amount nil @ %s for %s",
			t.Time.Format(coin.DateFormat),
			acc,
		)
	}
}

func (ts *totals) String() string {
	if ts == nil {
		return "nil()"
	}
	count := len(ts.all)
	if count == 0 {
		return "0()"
	}
	from := ts.all[0].Time.Format(coin.DateFormat)
	if count == 1 {
		return fmt.Sprintf("1(%s)", from)
	}
	to := ts.all[count-1].Time.Format(coin.DateFormat)
	return fmt.Sprintf("%d(%s-%s)", count, from, to)
}

type accountTotals map[*coin.Account]*totals

func (ats accountTotals) String() string {
	var items []string
	for acc, ts := range ats {
		n := "nil"
		if acc != nil {
			n = acc.Name
		}
		items = append(items, fmt.Sprintf("%s:%s", n, ts))
	}
	return fmt.Sprintf("totals{%s}", strings.Join(items, ", "))
}

func (ats accountTotals) newTotals(acc *coin.Account, by *reducer) *totals {
	ts := &totals{reducer: by}
	ats[acc] = ts
	return ts
}

func (ats accountTotals) widths(order []*coin.Account) (widths []int) {
	for _, acc := range order {
		widths = append(widths, ats[acc].maxWidth())
	}
	return widths
}

func (ats accountTotals) magnitudes() (magnitudes map[*coin.Account]*coin.Amount) {
	magnitudes = map[*coin.Account]*coin.Amount{}
	for acc, ts := range ats {
		magnitudes[acc] = ts.cumMagnitude()
	}
	return magnitudes
}

func (ats accountTotals) accounts() (accounts []*coin.Account) {
	for acc := range ats {
		accounts = append(accounts, acc)
	}
	return accounts
}

func (ats accountTotals) makeCumulative() {
	for _, ts := range ats {
		ts.makeCumulative()
	}
}

// top reduces the totals to top n by maximum magnitude + others (account == nil!)
func (ats accountTotals) top(n int) (topn accountTotals, order []*coin.Account) {
	topn = accountTotals{}
	magnitudes := ats.magnitudes()
	var accounts []*coin.Account
	for _, acc := range ats.accounts() {
		if magnitudes[acc] != nil {
			accounts = append(accounts, acc)
		}
	}

	sort.Slice(accounts, func(i int, j int) bool {
		return magnitudes[accounts[i]].IsBigger(magnitudes[accounts[j]])
	})
	if len(accounts) <= n {
		return ats, accounts
	}
	for _, acc := range accounts[:n] {
		topn[acc] = ats[acc]
	}
	rest := accounts[n:]
	accounts = accounts[:n]
	if len(rest) > 1 {
		others := ats[rest[0]]
		for _, acc := range rest[1:] {
			others.merge(ats[acc])
		}
		topn[nil] = others
		accounts = append(accounts, nil)
	}
	return topn, accounts
}

// mergeTime backfills of all totals with times from ts,
// so if ts has a union or superset of times of all the totals
// this will align them all.
func (ats accountTotals) mergeTime(ts *totals) {
	for _, ts2 := range ats {
		ts2.mergeTime(ts)
	}
}

func (ats accountTotals) validate() {
	for acc, ts := range ats {
		n := "nil"
		if acc != nil {
			n = acc.FullName
		}
		ts.validate(n)
	}
	fmt.Println(ats)
}

// santize removes accounts that don't have any totals
func (ats accountTotals) sanitize() {
	var empty []*coin.Account
	for acc, ts := range ats {
		if len(ts.all) == 0 {
			empty = append(empty, acc)
		}
	}
	for _, acc := range empty {
		delete(ats, acc)
	}
}

func (ats accountTotals) output(f io.Writer,
	order []*coin.Account,
	label func(*coin.Account) string,
	format string,
) {
	switch format {
	case "json":
		ats.rows(order, label).writeJSON(f)
	case "csv":
		ats.rows(order, label).writeCSV(f)
	default:
		ats.print(f, order, label)
	}
}

func (ats accountTotals) print(f io.Writer,
	order []*coin.Account,
	label func(*coin.Account) string,
) {
	firstCol := ats[order[0]]
	width1 := len(firstCol.all[0].Time.Format(firstCol.format))
	widths := ats.widths(order)
	format := []string{"%*s "}
	if label != nil {
		labels := make([]string, len(order))
		format2 := []string{"%*s "}
		for i, acc := range order {
			l := label(acc)
			labels[i] = l
			if len(l) > widths[i] {
				widths[i] = len(l)
			}
			format = append(format, " %*a ")
			format2 = append(format2, " %*s ")
		}
		args := []interface{}{width1, ""}
		for i := range widths {
			args = append(args, widths[i])
			args = append(args, labels[i])
		}
		fmtString := strings.TrimSpace(strings.Join(format2, "|")) + "\n"
		fmt.Fprintf(f, fmtString, args...)
	}
	fmtString := strings.TrimSpace(strings.Join(format, "|")) + "\n"
	for i := range firstCol.all {
		tm := firstCol.all[i].Time.Format(firstCol.format)
		args := []interface{}{width1, tm}
		for ii, acc := range order {
			ts := ats[acc]
			check.If(ts != nil, "nil totals for %s\n", label(acc))
			t := ts.all[i]
			tm2 := t.Time.Format(firstCol.format)
			check.If(tm == tm2, "%s[%d]: %s != %s\n", label(acc), i, tm, tm2)
			args = append(args, widths[ii])
			args = append(args, t.Amount)
		}
		fmt.Fprintf(f, fmtString, args...)
	}
}

func (ats accountTotals) rows(
	order []*coin.Account,
	label func(*coin.Account) string,
) (rs rows) {
	header := []string{"Date"}
	for _, acc := range order {
		header = append(header, label(acc))
	}
	rs = append(rs, header)
	firstCol := ats[order[0]]
	for i := range firstCol.all {
		tm := firstCol.all[i].Time.Format(firstCol.format)
		row := []string{tm}
		for _, acc := range order {
			t := ats[acc].all[i]
			tm2 := t.Time.Format(firstCol.format)
			check.If(tm == tm2, "%s[%d]: %s != %s\n", label(acc), i, tm, tm2)
			row = append(row, t.Amount.String())
		}
		rs = append(rs, row)
	}
	return rs
}

type rows [][]string

func (rs rows) writeCSV(f io.Writer) {
	w := csv.NewWriter(f)
	for _, r := range rs {
		w.Write(r)
	}
	w.Flush()
}

func (rs rows) writeJSON(f io.Writer) {
	w := json.NewEncoder(f)
	for _, r := range rs {
		w.Encode(r)
	}
}

// reducer coerces time to specified period
// and carries corresponding time format string.
type reducer struct {
	reduce func(t time.Time) time.Time
	format string
}

var week = reducer{
	reduce: func(t time.Time) time.Time {
		dow := int(t.Weekday())
		t = t.AddDate(0, 0, -dow)
		y, m, d := t.Date()
		return time.Date(y, m, d, 12, 0, 0, 0, time.UTC)
	},
	format: coin.DateFormat,
}

var month = reducer{
	reduce: func(t time.Time) time.Time {
		y, m, _ := t.Date()
		return time.Date(y, m, 1, 12, 0, 0, 0, time.UTC)
	},
	format: coin.MonthFormat,
}

var quarter = reducer{
	reduce: func(t time.Time) time.Time {
		y, m, _ := t.Date()
		m = ((m - 1) / 3 * 3) + 1
		return time.Date(y, m, 1, 12, 0, 0, 0, time.UTC)
	},
	format: coin.MonthFormat,
}

var year = reducer{
	reduce: func(t time.Time) time.Time {
		y, _, _ := t.Date()
		return time.Date(y, time.January, 1, 12, 0, 0, 0, time.UTC)
	},
	format: coin.YearFormat,
}
