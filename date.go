package coin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mkobetic/coin/rex"
)

func init() {
	var m time.Month
	Year, m, Day = time.Now().Date()
	Month = int(m)
}

// Cache today's values, so that we can spoof today for testing
var Year, Month, Day int

type Date struct {
	time.Time
}

func (d *Date) String() string {
	return d.Format(DateFormat)
}

func (d *Date) Set(s string) (err error) {
	match := DateREX.Match([]byte(s))
	if match == nil {
		return fmt.Errorf("Invalid date: %s", s)
	}
	d.Time, err = parseDate(match, 0)
	return err
}

var DateFormat = "2006/01/02"
var ymd = rex.MustCompile(`` +
	`((?P<ymd>((?P<ymdy>\d\d(\d\d)?)/)?(?P<ymdm>\d{1,2})/(?P<ymdd>\d{1,2}))|` +
	`(?P<ym>(?P<ymy>\d{4})(/(?P<ymm>\d{1,2}))?))`)
var offset = rex.MustCompile(`(?P<offset>[+-]\d+[d|w|m|y])`)
var DateREX = rex.MustCompile(`(?P<date>%s?%s?)`, ymd, offset)

func mustParseDate(match map[string]string, idx int) time.Time {
	d, err := parseDate(match, idx)
	if err != nil {
		panic(err)
	}
	return d
}

func parseDate(match map[string]string, idx int) (t time.Time, err error) {
	if idx > 0 {
		return t, fmt.Errorf("Multiple date fields not implemented!")
	}
	// Set date to today
	y, m, d := Year, Month, Day
	offset := match["offset"]
	if match["ymd"] != "" {
		d, _ = strconv.Atoi(match["ymdd"])
		mm, _ := strconv.Atoi(match["ymdm"])
		if yy := match["ymdy"]; yy != "" {
			yyy, _ := strconv.Atoi(yy)
			if yyy < 100 {
				yyy = y/1000*1000 + yyy
				if yyy < y && y-yyy > 50 {
					yyy += 100
				} else if yyy > y && yyy-y > 50 {
					yyy -= 100
				}
				y = yyy
			} else {
				y = yyy
			}
		} else {
			if mm < m && m-mm > 6 {
				y += 1
			} else if m < mm && mm-m > 6 {
				y -= 1
			}
		}
		m = mm
	} else if match["ym"] != "" {
		d, m = 1, 1
		y, _ = strconv.Atoi(match["ymy"])
		if mm := match["ymm"]; mm != "" {
			m, _ = strconv.Atoi(mm)
		}
	} else if offset == "" {
		return t, fmt.Errorf("no match for date: %v", match)
	}

	t = time.Date(y, time.Month(m), d, 12, 0, 0, 0, time.UTC)
	if offset != "" {
		e := len(offset) - 1
		o, _ := strconv.Atoi(offset[:e])
		switch offset[e] {
		case 'd':
			t = t.AddDate(0, 0, o)
		case 'w':
			t = t.AddDate(0, 0, o*7)
		case 'm':
			t = t.AddDate(0, o, 0)
		case 'y':
			t = t.AddDate(o, 0, 0)
		}
	}
	return t, nil
}
