package coin

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/mkobetic/coin/rex"
)

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
var DateREX = rex.MustCompile(`(?P<date>\d\d\d\d/\d\d/\d\d)`)

func mustParseDate(match map[string]string, idx int) time.Time {
	t, err := time.Parse(DateFormat, match["date"])
	if err != nil {
		panic(err)
	}
	return t
}
