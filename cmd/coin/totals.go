package main

import (
	"fmt"
	"io"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/mkobetic/coin"
)

type total struct {
	time.Time
	*coin.Amount
}

type totals struct {
	// this must be set before using
	by func(time.Time) time.Time
	// these are internal
	all     []*total
	current *total
}

// add amount at time t, t must not be before ts.current.Time,
// i.e. items being added must be sorted by time.
func (ts *totals) add(t time.Time, a *coin.Amount) {
	period := ts.by(t)
	if ts.current != nil && ts.current.Equal(period) {
		ts.current.AddIn(a)
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
		ts.add(t.Time, coin.NewAmount(new(big.Int), t.Commodity))
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

// cumulative converts ts to a cumulative totals sequence
func (ts *totals) cumulative() {
	if ts.current == nil {
		return
	}
	all := ts.all
	cum := coin.NewAmount(new(big.Int), ts.current.Commodity)
	ts.all, ts.current = nil, nil
	for _, t := range all {
		cum.AddIn(t.Amount)
		ts.add(t.Time, cum)
	}
}

type accountTotals map[*coin.Account]*totals

func (ats accountTotals) newTotals(acc *coin.Account, by func(time.Time) time.Time, cumulative bool) *totals {
	ts := &totals{by: by}
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
		magnitudes[acc] = ts.maxMagnitude()
	}
	return magnitudes
}

func (ats accountTotals) accounts() (accounts []*coin.Account) {
	for acc := range ats {
		accounts = append(accounts, acc)
	}
	return accounts
}

func (ats accountTotals) cumulative() {
	for _, ts := range ats {
		ts.cumulative()
	}
}

// top reduces the totals to top n by maximum magnitude + others (key: nil)
func (ats accountTotals) top(n int) (topn accountTotals, order []*coin.Account) {
	topn = accountTotals{}
	magnitudes := ats.magnitudes()
	accounts := ats.accounts()
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
	if len(rest) > 1 {
		others := ats[rest[0]]
		for _, acc := range rest[1:] {
			others.merge(ats[acc])
		}
		topn[nil] = others
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

func (ats accountTotals) print(f io.Writer,
	order []*coin.Account,
	label func(*coin.Account) string,
	dateFmt string,
) {
	firstCol := ats[order[0]].all
	width1 := len(firstCol[0].Time.Format(dateFmt))
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
	for i := range firstCol {
		args := []interface{}{
			width1,
			firstCol[i].Time.Format(dateFmt),
		}
		for ii, acc := range order {
			args = append(args, widths[ii])
			args = append(args, ats[acc].all[i].Amount)
		}
		fmt.Fprintf(f, fmtString, args...)
	}
}

func month(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 12, 0, 0, 0, time.UTC)
}

func year(t time.Time) time.Time {
	y, _, _ := t.Date()
	return time.Date(y, time.January, 1, 12, 0, 0, 0, time.UTC)
}

func week(t time.Time) time.Time {
	dow := int(t.Weekday())
	t = t.AddDate(0, 0, -dow)
	y, m, d := t.Date()
	return time.Date(y, m, d, 12, 0, 0, 0, time.UTC)
}
