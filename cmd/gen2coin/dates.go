package main

/*

	Recuring date sequence generators

See also:
* https://martinfowler.com/apsupp/recurring.pdf
* iCal RRULE rule: https://www.kanzaki.com/docs/ical/recur.html
* https://learn.microsoft.com/en-us/graph/outlook-schedule-recurring-events

*/

import (
	"sort"
	"time"
)

type dateGen func(begin, end time.Time) []time.Time

var anyday = []time.Weekday{time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday}
var weekday = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}
var weekend = []time.Weekday{time.Sunday, time.Saturday}

// weekly generates list of dates (sorted ascendigly) for @count out of @days every @interval weeks.
// if @count is less then len(@days), the days are picked randomly.
func weekly(count int, interval int, days ...time.Weekday) dateGen {
	return func(begin, end time.Time) (dates []time.Time) {
		current := begin.AddDate(0, 0, int(-begin.Weekday()))
		if len(days) == 0 {
			days = anyday
		}
		if interval < 1 {
			interval = 1
		}
		for current.Before(end) {
			picks := days
			if count < len(picks) {
				// random picks, each day can only be picked once
				var candidates []time.Weekday
				candidates = append(candidates, picks...)
				picks = nil
				for i := 0; i < count && len(candidates) > 0; i++ {
					offset := rnd.Intn(len(candidates))
					pick := candidates[offset]
					// drop the pick from candidates
					copy(candidates[offset:], candidates[offset+1:])
					candidates = candidates[:len(candidates)-1]
					picks = append(picks, pick)
				}
				sort.Slice(picks, func(i, j int) bool { return picks[i] < picks[j] })
			}
			for _, pick := range picks {
				pick := current.AddDate(0, 0, int(pick))
				if !(end.Before(pick) || pick.Before(begin)) {
					dates = append(dates, pick)
				}
			}
			current = current.AddDate(0, 0, 7*interval)
		}
		return dates
	}
}

var mDays = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}

// monthly generates ascending list of dates for @count out of @days every @interval months.
// @days are days of month starting at 1, if negative counted from the end of the month,
// 0 and other invalid days (for given month) are ignored.
func monthly(count int, interval int, days ...int) dateGen {
	return func(begin, end time.Time) (dates []time.Time) {
		current := begin.AddDate(0, 0, 1-begin.Day())
		if interval < 1 {
			interval = 1
		}
		for current.Before(end) {
			picks := days
			if len(picks) == 0 {
				_, _, nDays := current.AddDate(0, 1, -1).Date()
				picks = mDays[:nDays]
			}
			if count < len(picks) {
				// random picks, each day can only be picked once
				var candidates []int
				candidates = append(candidates, picks...)
				picks = nil
				for i := 0; i < count && len(candidates) > 0; i++ {
					offset := rnd.Intn(len(candidates))
					pick := candidates[offset]
					// drop the pick from candidates
					copy(candidates[offset:], candidates[offset+1:])
					candidates = candidates[:len(candidates)-1]
					picks = append(picks, pick)
				}
			}
			var newDates []time.Time
			cY, cM, _ := current.Date()
			for _, p := range picks {
				var pick time.Time
				if p <= 0 {
					pick = current.AddDate(0, 1, p)
				} else {
					pick = current.AddDate(0, 0, p-1)
				}
				pY, pM, _ := pick.Date()
				if cY == pY && cM == pM && !(end.Before(pick) || pick.Before(begin)) {
					newDates = append(newDates, pick)
				}
			}
			sort.Slice(newDates, func(i, j int) bool { return newDates[i].Before(newDates[j]) })
			dates = append(dates, newDates...)
			current = current.AddDate(0, 1*interval, 0)
		}
		return dates
	}
}
