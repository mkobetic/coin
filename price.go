package coin

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mkobetic/coin/rex"
)

type Price struct {
	Commodity *Commodity
	Currency  *Commodity
	Value     *Amount
	Time      time.Time

	CommodityId string
	currencyId  string
	line        uint
	file        string
}

var Prices []*Price

func (p *Price) Write(w io.Writer, ledger bool) error {
	date := p.Time.Format(DateFormat)
	_, err := io.WriteString(w, "P "+date+" "+p.Commodity.SafeId(ledger)+" ")
	if err != nil {
		return err
	}
	err = p.Value.Write(w, ledger)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, "\n")
	return err
}

var priceREX = rex.MustCompile(`P %s\s+%s\s+%s`, DateREX, CommodityREX, AmountREX)

func (p *Parser) parsePrice(fn string) (*Price, error) {
	match := priceREX.Match(p.Bytes())
	if match == nil {
		return nil, fmt.Errorf("invalid price line")
	}
	date := mustParseDate(match, 0)
	currencyId := string(match["commodity2"])
	line := p.lineNr
	location := fmt.Sprintf("%s:%d", fn, line)
	c := MustFindCommodity(currencyId, location)
	amt, err := parseAmount(match["amount"], c)
	if err != nil {
		return nil, err
	}
	commodityId := string(match["commodity1"])
	p.Scan() // advance to next line before returning
	return &Price{
		Time:        date,
		Value:       amt,
		CommodityId: commodityId,
		currencyId:  currencyId,
		line:        line,
		file:        fn,
	}, nil
}

func (p *Price) String() string {
	var b strings.Builder
	p.Write(&b, false)
	return b.String()
}

func (p *Price) Location() string {
	return fmt.Sprintf("%s:%d", p.file, p.line)
}

func (p *Price) MarshalJSON() ([]byte, error) {
	value := map[string]interface{}{
		"time":      p.Time.Format(DateFormat),
		"commodity": p.Commodity.Id,
		"currency":  p.Currency.Id,
		"value":     p.Value,
		"location":  p.Location(),
	}
	return json.MarshalIndent(value, "", "\t")
}
