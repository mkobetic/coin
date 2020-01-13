package main

import (
	"math/big"
	"time"

	"github.com/mkobetic/coin"
)

type total struct {
	time.Time
	*coin.Amount
}

func newTotal(t time.Time, c *coin.Commodity) *total {
	return &total{
		Time:   t,
		Amount: coin.NewAmount(big.NewInt(0), c),
	}
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
