package coin

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"time"
)

type Parser struct {
	*bufio.Scanner
}

type Item interface {
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		Scanner: bufio.NewScanner(r),
	}
}

func (p *Parser) Next() (Item, error) {
	if !p.Scan() {
		return nil, p.Err()
	}
	switch line := p.Bytes(); {
	case len(bytes.TrimSpace(line)) == 0:
		return p.Next()
	case bytes.ContainsAny(line[:1], ";#%|*"):
		return p.Next()
	case bytes.HasPrefix(line, []byte("account")):
		return p.parseAccount()
	case bytes.HasPrefix(line, []byte("commodity ")):
		return p.parseCommodity()
	case bytes.HasPrefix(line, []byte("test ")):
		return p.parseTest()
	case bytes.HasPrefix(line, []byte("P ")):
		return p.parsePrice()
	case '0' <= line[0] && line[0] <= '9':
		return p.parseTransaction()
	default:
		return nil, fmt.Errorf("Unrecognized item: %s", line)
	}
}

var DateFormat = "2006/01/02"
var DateRE = `(\d\d\d\d/\d\d/\d\d)`

func mustParseDate(ts []byte) time.Time {
	t, err := time.Parse(DateFormat, string(ts))
	if err != nil {
		panic(err)
	}
	return t
}
