package main

import (
	"strings"
	"testing"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/assert"
)

var CAD = &coin.Commodity{Id: "CAD", Decimals: 2}

func makeTimeTotals(by *timeReducer, totals ...string) *timeTotals {
	var all []*timeTotal
	var current *timeTotal
	for _, in := range totals {
		parts := strings.Split(in, ":")
		t := coin.MustParseDate(parts[0])
		var amt *coin.Amount
		if len(parts) > 1 {
			amt = coin.MustParseAmount(parts[1], CAD)
		} else {
			amt = coin.NewZeroAmount(CAD)
		}
		current = &timeTotal{
			Time:  t,
			total: amt,
		}
		all = append(all, current)
	}
	return &timeTotals{
		timeReducer: by,
		all:         all,
		current:     current,
	}
}

func Test_MergeTotals(t *testing.T) {
	coin.WithYear(2025, func() {
		t1 := makeTimeTotals(&month, "1/1:1", "2/2:2", "4/4:4")
		t2 := makeTimeTotals(&month, "2/2:10", "3/3:30", "5/5:50")
		assert.Equal(t, t1.String(), "3(2025/01/01-2025/04/04)")
		assert.Equal(t, t2.String(), "3(2025/02/02-2025/05/05)")
		t1.mergeTotals(t2, false)
		assert.EqualStrings(t,
			tsToStrings(t1.all...),
			"2025/01/01: 1.00",
			"2025/02/02: 12.00",
			"2025/03/03: 30.00",
			"2025/04/04: 4.00",
			"2025/05/05: 50.00",
		)
	})
}

func Test_MergeTotalsTimeOnly(t *testing.T) {
	coin.WithYear(2025, func() {
		t1 := makeTimeTotals(&month, "1/1:1", "2/2:2", "4/4:4")
		t2 := makeTimeTotals(&month, "2/2:10", "3/3:30", "5/5:50")
		assert.Equal(t, t1.String(), "3(2025/01/01-2025/04/04)")
		assert.Equal(t, t2.String(), "3(2025/02/02-2025/05/05)")
		t1.mergeTotals(t2, true)
		assert.EqualStrings(t,
			tsToStrings(t1.all...),
			"2025/01/01: 1.00",
			"2025/02/02: 2.00",
			"2025/03/03: 0.00",
			"2025/04/04: 4.00",
			"2025/05/05: 0.00",
		)
	})
}
