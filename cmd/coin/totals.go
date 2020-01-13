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
