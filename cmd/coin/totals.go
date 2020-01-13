package main

import (
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

func (ts *totals) add(t time.Time, a *coin.Amount) {
	period := ts.by(t)
	if ts.current != nil && ts.current.Equal(period) {
		ts.current.AddIn(a)
		return
	}
	ts.current = &total{Time: period, Amount: a.Copy()}
	ts.all = append(ts.all, ts.current)
}

func (ts *totals) addTotals(ts2 ...*total) {
	for _, t := range ts2 {
		ts.add(t.Time, t.Amount)
	}
}

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
