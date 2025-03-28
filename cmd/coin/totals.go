package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

// timeTotal represents a total amount for a time period.
// optionally
type timeTotal struct {
	time.Time // time is zero if we're not reducing by time
	// exactly one of the following two must be not nil
	total  *coin.Amount            // total if we are not reducing by other criteria
	totals map[string]*coin.Amount // totals by other criteria
}

func (t *timeTotal) Categories() (keys []string) {
	for k := range t.totals {
		keys = append(keys, k)
	}
	return keys
}

func (t *timeTotal) AddIn(t2 *timeTotal, keysOnly bool) {
	check.If(t.Time.Equal(t2.Time), "%v is not equal to %v", t.Time, t2.Time)
	if t.total != nil && !keysOnly {
		t.total.AddIn(t2.total)
		return
	}
	for k, v := range t2.totals {
		amt := t.totals[k]
		if amt != nil {
			if !keysOnly {
				amt.AddIn(v)
			}
		} else {
			if keysOnly {
				t.totals[k] = coin.NewZeroAmount(v.Commodity)
			} else {
				t.totals[k] = v.Copy()
			}
		}
	}
}

func (t *timeTotal) Copy(keysOnly bool) *timeTotal {
	if t.total != nil {
		return &timeTotal{Time: t.Time, total: t.total.Copy()}
	}
	totals := make(map[string]*coin.Amount)
	for k, v := range t.totals {
		if keysOnly {
			totals[k] = coin.NewZeroAmount(v.Commodity)
		} else {
			totals[k] = v.Copy()
		}
	}
	return &timeTotal{Time: t.Time, totals: totals}
}

func (t *timeTotal) Commodity() *coin.Commodity {
	if t.total != nil {
		return t.total.Commodity
	}
	for _, v := range t.totals {
		return v.Commodity
	}
	return nil
}

func (t *timeTotal) String() string {
	output := t.Time.Format(coin.DateFormat) + ": "
	if t.total != nil {
		return output + t.total.String()
	}
	var totals []string
	for k, v := range t.totals {
		totals = append(totals, fmt.Sprintf("%s: %s", k, v))
	}
	return output + "{ " + strings.Join(totals, ", ") + " }"
}

func tsToStrings(ts ...*timeTotal) (out []string) {
	for _, t := range ts {
		out = append(out, t.String())
	}
	return out
}

// timeTotals aggregates time series amounts for accounts.
// It must be initialized with a reducer before use.
type timeTotals struct {
	// at least one of these must be set before using
	*timeReducer
	*categoryReducer
	// these are internal
	all     []*timeTotal // ordered totals by time if we are reducing by time
	current *timeTotal   // latest total, if not reducing by time then the only total
}

func (ts *timeTotals) Commodity() *coin.Commodity {
	if ts.current != nil {
		if c := ts.current.Commodity(); c != nil {
			return c
		}
	}
	for _, v := range ts.all {
		if c := v.Commodity(); c != nil {
			return c
		}
	}
	return nil
}

// add posting to the totals
// Assumes postings are added in time order,
// which should be true since account postings are sorted by time.
func (ts *timeTotals) add(p *coin.Posting) {
	var period time.Time
	if ts.timeReducer != nil {
		period = ts.timeReducer.reduce(p.Transaction.Posted)
	}
	if ts.current != nil && ts.current.Equal(period) {
		if ts.categoryReducer != nil {
			category := ts.categoryReducer.reduce(p)
			amt := ts.current.totals[category]
			if amt != nil {
				check.NoError(amt.AddIn(p.Quantity), "cannot add %a to totals", p.Quantity)
			} else {
				ts.current.totals[category] = p.Quantity.Copy()
			}
		} else {
			check.NoError(ts.current.total.AddIn(p.Quantity), "cannot add %a to totals", p.Quantity)
		}
		return
	}
	amt := p.Quantity.Copy()
	if ts.categoryReducer != nil {
		category := ts.categoryReducer.reduce(p)
		ts.current = &timeTotal{Time: period, totals: map[string]*coin.Amount{category: amt}}
	} else {
		ts.current = &timeTotal{Time: period, total: amt}
	}
	if ts.timeReducer != nil {
		ts.all = append(ts.all, ts.current)
	}
}

func (ts *timeTotals) newTotal(t time.Time, c *coin.Commodity) *timeTotal {
	if ts.categoryReducer != nil {
		return &timeTotal{Time: t, totals: make(map[string]*coin.Amount)}
	} else {
		check.If(c != nil, "need commodity to create total")
		return &timeTotal{Time: t, total: coin.NewZeroAmount(c)}
	}
}

// mergeTotals merges times and categories from ts2 into ts.
// keysOnly = false only add times/categories that are missing with zero value
// keysOnly = true add the ts2 amounts to corresponding ts values.
func (ts *timeTotals) mergeTotals(ts2 *timeTotals, keysOnly bool) {
	if ts.timeReducer == nil {
		if ts2.current == nil {
			return
		}
		if ts.current == nil {
			ts.current = ts2.current.Copy(keysOnly)
		} else {
			ts.current.AddIn(ts2.current, keysOnly)
		}
		return
	}
	var offsets []int
	var extras []*timeTotal
	remaining := ts.all
	// Add totals with matching time into ts,
	// gather non-matching into extras with offsets where they should go
	var j int
	for _, t := range ts2.all {
		i := sort.Search(len(remaining), func(i int) bool {
			return !remaining[i].Time.Before(t.Time)
		})
		remaining = remaining[i:]
		j += i
		if len(remaining) > 0 && remaining[0].Time.Equal(t.Time) {
			remaining[0].AddIn(t, keysOnly)
		} else {
			offsets = append(offsets, j)
			j = 0
			extras = append(extras, t)
		}
	}
	if len(extras) == 0 {
		// if everything matched we are done
		return
	}
	// Otherwise inject the extras into the timeline.
	remaining = ts.all
	ts.all = make([]*timeTotal, 0, len(ts.all)+len(extras))
	for i := 0; i < len(extras); i++ {
		mark := offsets[i]
		extra := extras[i]
		com := ts.Commodity()
		if com == nil {
			com = extra.Commodity()
		}
		newExtra := ts.newTotal(extra.Time, com)
		newExtra.AddIn(extra, keysOnly)
		ts.all = append(ts.all, remaining[:mark]...)
		ts.all = append(ts.all, newExtra)
		remaining = remaining[mark:]
	}
	ts.all = append(ts.all, remaining...)
	ts.current = ts.all[len(ts.all)-1]
}

// merge two sorted totals
func (ts *timeTotals) merge(ts2 *timeTotals) {
	ts.mergeTotals(ts2, false)
}

// mergeKeys backfills ts with missing times and categories from ts2
func (ts *timeTotals) mergeKeys(ts2 *timeTotals) {
	ts.mergeTotals(ts2, true)
}

// maxWidth returns the largest amount width needed for printing
func (ts *timeTotals) maxWidth() int {
	if ts == nil || ts.current == nil {
		return 0
	}
	dec := ts.Commodity().Decimals
	var max int
	if len(ts.all) == 0 {
		for _, amt := range ts.current.totals {
			if w := amt.Width(dec); w > max {
				max = w
			}
		}
	} else {
		for _, t := range ts.all {
			if t.total == nil {
				for _, v := range t.totals {
					if w := v.Width(dec); w > max {
						max = w
					}
				}
			} else {
				if w := t.total.Width(dec); w > max {
					max = w
				}
			}
		}
	}
	return max
}

func (ts *timeTotals) maxPeriodWidth() int {
	if ts.timeReducer == nil {
		return 0
	}
	return len(ts.all[0].Time.Format(ts.format))
}

func (ts *timeTotals) maxCategoryWidth() int {
	periods := ts.all
	if len(periods) == 0 {
		periods = append(periods, ts.current)
	}
	max := 0
	for _, period := range periods {
		for k := range period.totals {
			if len(k) > max {
				max = len(k)
			}
		}
	}
	return max
}

// cumMagnitude returns the sum of the totals
func (ts *timeTotals) cumMagnitude() *coin.Amount {
	if ts == nil || ts.current == nil {
		return nil
	}
	total := coin.NewZeroAmount(ts.Commodity())
	if len(ts.all) == 0 {
		for _, amt := range ts.current.totals {
			check.NoError(total.AddIn(amt), "computing cumulative magnitude")
		}
	} else {
		for _, t := range ts.all {
			if t.total != nil {
				check.NoError(total.AddIn(t.total), "computing cumulative magnitude")
			} else {
				for _, v := range t.totals {
					check.NoError(total.AddIn(v), "computing cumulative magnitude")
				}
			}
		}
	}
	return total
}

// cumulative converts ts to a cumulative totals sequence
func (ts *timeTotals) makeCumulative() {
	check.If(ts.categoryReducer == nil, "category aggregation cannot be cumulative")
	if ts.current == nil {
		return
	}
	cum := coin.NewZeroAmount(ts.Commodity())
	for _, t := range ts.all {
		err := cum.AddIn(t.total)
		check.NoError(err, "converting totals to cumulative")
		t.total = cum.Copy()
	}
}

func (ts *timeTotals) validate(acc string) {
	check.If(ts.current != nil, "current nil for %s", acc)
	for _, t := range ts.all {
		check.If(t.total != nil, "amount nil @ %s for %s",
			t.Time.Format(coin.DateFormat),
			acc,
		)
	}
}

func (ts *timeTotals) String() string {
	if ts == nil || ts.current == nil {
		return "nil()"
	}
	count := len(ts.all)
	if count == 0 {
		count = len(ts.current.totals)
		var keys []string
		var kcount = 0
		for k := range ts.current.totals {
			keys = append(keys, k)
			if kcount > 2 {
				break
			}
		}
		return fmt.Sprintf("%d(%s ...)", count, strings.Join(keys, ", "))
	}
	from := ts.all[0].Time.Format(coin.DateFormat)
	if count == 1 {
		return fmt.Sprintf("1(%s)", from)
	}
	to := ts.all[count-1].Time.Format(coin.DateFormat)
	return fmt.Sprintf("%d(%s-%s)", count, from, to)
}

// accountTotals represents timeTotals across a hierarchy of accounts.
// allows aggregated register to operate on the entire result.
type accountTotals map[*coin.Account]*timeTotals

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

func (ats accountTotals) newTotals(acc *coin.Account, period *timeReducer, category *categoryReducer) *timeTotals {
	ts := &timeTotals{timeReducer: period, categoryReducer: category}
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
func (ats accountTotals) mergeTime(ts *timeTotals) {
	for _, ts2 := range ats {
		if ts2 != ts {
			ts2.mergeKeys(ts)
		}
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
		if ts.current == nil {
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
	format := []string{"%*s "}
	periodWidth := firstCol.maxPeriodWidth()
	categoryWidth := firstCol.maxCategoryWidth()
	if periodWidth > 0 && categoryWidth > 0 {
		format = append(format, " %*s ")
	}
	widths := ats.widths(order)
	if label != nil {
		format2 := append([]string{}, format...)
		labels := make([]string, len(order))
		for i, acc := range order {
			l := label(acc)
			labels[i] = l
			if len(l) > widths[i] {
				widths[i] = len(l)
			}
			format = append(format, " %*a ")
			format2 = append(format2, " %*s ")
		}
		var args []interface{}
		if periodWidth > 0 {
			args = append(args, periodWidth, "")
		}
		if categoryWidth > 0 {
			args = append(args, categoryWidth, "")
		}
		for i := range widths {
			args = append(args, widths[i], labels[i])
		}
		fmtString := strings.TrimSpace(strings.Join(format2, "|")) + "\n"
		fmt.Fprintf(f, fmtString, args...)
	}
	fmtString := strings.TrimSpace(strings.Join(format, "|")) + "\n"
	periods := firstCol.all
	if len(periods) == 0 {
		periods = append(periods, firstCol.current)
	}
	for i, period := range periods {
		categories := period.Categories()
		if len(categories) == 0 {
			categories = []string{""}
		} else {
			slices.Sort(categories)
		}
		for _, category := range categories {
			var args []interface{}
			if periodWidth > 0 {
				tm := period.Time.Format(firstCol.format)
				args = append(args, periodWidth, tm)
			}
			if categoryWidth > 0 {
				args = append(args, categoryWidth, category)
			}
			for ii, acc := range order {
				ts := ats[acc]
				check.If(ts != nil, "nil totals for %s\n", label(acc))
				t := ts.current
				if periodWidth > 0 {
					t = ts.all[i]
					check.If(period.Time.Equal(t.Time), "%s[%d]: %s != %s\n", label(acc), i,
						t.Time.Format(firstCol.format),
						period.Time.Format(firstCol.format))
				}
				amt := t.total
				if category != "" {
					amt = t.totals[category]
				}
				args = append(args, widths[ii], amt)
			}
			fmt.Fprintf(f, fmtString, args...)
		}
	}
}

// rows converts accountTotals to a plain row result
// suitable for output as csv or json.
func (ats accountTotals) rows(
	order []*coin.Account,
	label func(*coin.Account) string,
) (rs rows) {
	var header []string
	firstCol := ats[order[0]]
	hasPeriod := firstCol.timeReducer != nil
	hasCategory := firstCol.categoryReducer != nil
	if hasPeriod {
		header = append(header, "Date")
	}
	if hasCategory {
		header = append(header, "Category")
	}
	for _, acc := range order {
		header = append(header, label(acc))
	}
	rs = append(rs, header)
	periods := firstCol.all
	if len(periods) == 0 {
		periods = append(periods, firstCol.current)
	}
	for i, period := range periods {
		categories := period.Categories()
		if len(categories) == 0 {
			categories = []string{""}
		} else {
			slices.Sort(categories)
		}
		for _, category := range categories {
			var row []string
			if hasPeriod {
				row = append(row, period.Time.Format(firstCol.format))
			}
			if hasCategory {
				row = append(row, category)
			}
			for _, acc := range order {
				ts := ats[acc]
				check.If(ts != nil, "nil totals for %s\n", label(acc))
				t := ts.current
				if hasPeriod {
					t = ts.all[i]
					check.If(period.Time.Equal(t.Time), "%s[%d]: %s != %s\n", label(acc), i,
						t.Time.Format(firstCol.format),
						period.Time.Format(firstCol.format))
				}
				amt := t.total
				if category != "" {
					amt = t.totals[category]
				}
				row = append(row, amt.String())
			}
			rs = append(rs, row)
		}
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

// categoryReducer extracts category string from a posting
type categoryReducer struct {
	reduce func(p *coin.Posting) string
}

var payees = categoryReducer{
	reduce: func(p *coin.Posting) string {
		return p.Transaction.Description
	},
}

var tags = categoryReducer{
	reduce: func(p *coin.Posting) string {
		tags := p.Transaction.Tags.With(p.Tags)
		if len(tags) == 0 {
			return "<no-tags>"
		} else {
			return strings.Join(tags.KeysAndValues(), ",")
		}
	},
}

// timeReducer coerces time to specified period
// and carries corresponding time format string.
type timeReducer struct {
	reduce func(t time.Time) time.Time
	format string
}

var week = timeReducer{
	reduce: func(t time.Time) time.Time {
		dow := int(t.Weekday())
		t = t.AddDate(0, 0, -dow)
		y, m, d := t.Date()
		return time.Date(y, m, d, 12, 0, 0, 0, time.UTC)
	},
	format: coin.DateFormat,
}

var month = timeReducer{
	reduce: func(t time.Time) time.Time {
		y, m, _ := t.Date()
		return time.Date(y, m, 1, 12, 0, 0, 0, time.UTC)
	},
	format: coin.MonthFormat,
}

var quarter = timeReducer{
	reduce: func(t time.Time) time.Time {
		y, m, _ := t.Date()
		m = ((m - 1) / 3 * 3) + 1
		return time.Date(y, m, 1, 12, 0, 0, 0, time.UTC)
	},
	format: coin.MonthFormat,
}

var year = timeReducer{
	reduce: func(t time.Time) time.Time {
		y, _, _ := t.Date()
		return time.Date(y, time.January, 1, 12, 0, 0, 0, time.UTC)
	},
	format: coin.YearFormat,
}
