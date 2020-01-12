package coin

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/mkobetic/coin/rex"
)

func init() {
	var m time.Month
	Year, m, Day = time.Now().Date()
	Month = int(m)
}

type Parser struct {
	*bufio.Scanner
	finished bool
	lineNr   uint
}

type Item interface {
}

func NewParser(r io.Reader) *Parser {
	p := &Parser{Scanner: bufio.NewScanner(r)}
	p.Scan()
	return p
}

func (p *Parser) Scan() bool {
	p.finished = !p.Scanner.Scan()
	if !p.finished {
		p.lineNr++
	}
	return !p.finished
}

func (p *Parser) Next(fn string) (Item, error) {
	if p.finished {
		return nil, p.Err()
	}
	switch line := p.Bytes(); {
	case len(bytes.TrimSpace(line)) == 0:
		p.Scan()
		return p.Next(fn)
	case bytes.ContainsAny(line[:1], ";#%|*"):
		p.Scan()
		return p.Next(fn)
	case bytes.HasPrefix(line, []byte("account")):
		return p.parseAccount(fn)
	case bytes.HasPrefix(line, []byte("commodity ")):
		return p.parseCommodity(fn)
	case bytes.HasPrefix(line, []byte("test ")):
		return p.parseTest(fn)
	case bytes.HasPrefix(line, []byte("P ")):
		return p.parsePrice(fn)
	case '0' <= line[0] && line[0] <= '9':
		return p.parseTransaction(fn)
	default:
		return nil, fmt.Errorf("Unrecognized item: %s", line)
	}
}

var DateFormat = "2006/01/02"
var ymd = rex.MustCompile(`` +
	`((?P<ymd>((?P<year>\d\d(\d\d)?)/)?(?P<month>\d{1,2})/(?P<day>\d{1,2}))|` +
	`(?P<ym>(?P<ymy>\d{4})(/(?P<ymm>\d{1,2}))?))`)
var offset = rex.MustCompile(`(?P<offset>[+-]\d+[d|w|m|y])`)
var DateREX = rex.MustCompile(`(?P<date>%s?%s?)`, ymd, offset)

// Cache today's values, so that we can spoof today for testing
var Year, Month, Day int

func mustParseDate(match map[string]string, idx int) time.Time {
	if idx > 0 {
		panic("Multiple date fields not implemented!")
	}
	var t time.Time
	offset := match["offset"]
	y, m, d := Year, Month, Day
	if match["ymd"] != "" {
		d, _ = strconv.Atoi(match["day"])
		mm, _ := strconv.Atoi(match["month"])
		if yy := match["year"]; yy != "" {
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
		panic(fmt.Errorf("no match for date: %v", match))
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
	return t
}
