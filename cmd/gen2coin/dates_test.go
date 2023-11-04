package main

import (
	"strings"
	"testing"
	"time"

	"github.com/mkobetic/coin/assert"
)

func d(s string) time.Time {
	t, err := time.Parse("06/01/02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func Test_DateGen(t *testing.T) {
	for i, tc := range []struct {
		name     string
		from, to string
		gen      dateGen
		exp      string
	}{
		{
			"every Thursday",
			"20/12/07", "21/01/13",
			weekly(1, 1, time.Thursday),
			"20/12/10 Thu, 20/12/17 Thu, 20/12/24 Thu, 20/12/31 Thu, 21/01/07 Thu",
		},
		{
			"Tuesday and Friday every other week",
			"20/12/17", "21/01/13",
			weekly(2, 2, time.Tuesday, time.Friday),
			"20/12/18 Fri, 20/12/29 Tue, 21/01/01 Fri, 21/01/12 Tue",
		},
		{
			"10th, 20th and 30th of every month",
			"20/12/15", "21/03/15",
			monthly(3, 1, 10, 20, 30),
			"20/12/20 Sun, 20/12/30 Wed, 21/01/10 Sun, 21/01/20 Wed, 21/01/30 Sat, 21/02/10 Wed, 21/02/20 Sat, 21/03/10 Wed",
		},
		{
			"15th and 30th last day of every other month",
			"20/10/10", "21/06/10",
			monthly(2, 2, -15, -30),
			"20/10/17 Sat, 20/12/02 Wed, 20/12/17 Thu, 21/02/14 Sun, 21/04/01 Thu, 21/04/16 Fri, 21/06/01 Tue",
		},
	} {
		var b strings.Builder
		for i, d := range tc.gen(d(tc.from), d(tc.to)) {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(d.Format("06/01/02 Mon"))
		}
		assert.Equal(t, b.String(), tc.exp, "%d: %s", i, tc.name)
	}
}
