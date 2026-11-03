package coin

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"

	"github.com/mkobetic/coin/rex"
)

var AmountREX = rex.MustCompile(`(?P<amount>-?[\d]+(,\d{3})*(?P<decimals>\.[\d]+)?)\s+%s`, CommodityREX)

type Amount struct {
	*big.Int
	*Commodity
}

func NewAmountFrac(num, den *big.Int, c *Commodity) *Amount {
	b := new(big.Int).Set(num)
	if c.Decimals > 0 {
		b.Mul(b, bigPow10(c.Decimals))
	}
	b.Div(b, den)
	return NewAmount(b, c)
}

func NewZeroAmount(c *Commodity) *Amount {
	return NewAmount(new(big.Int), c)
}

func NewAmount(b *big.Int, c *Commodity) *Amount {
	return &Amount{b, c}
}

func (a *Amount) Copy() *Amount {
	return NewAmount(new(big.Int).Set(a.Int), a.Commodity)
}

func (a *Amount) String() string {
	val := new(big.Rat)
	val.SetFrac(a.Int, bigPow10(a.Commodity.Decimals))
	return val.FloatString(a.Commodity.Decimals)
}

// Format implements fmt.Formatter
func (a *Amount) Format(f fmt.State, c rune) {
	if f.Flag('#') {
		fmt.Fprintf(f, "%d/(10^%d) %s", a.Int, a.Decimals, a.Commodity.Id)
		return
	}
	p, ok := f.Precision()
	if !ok {
		p = a.Commodity.Decimals
	}
	val := new(big.Rat)
	val.SetFrac(a.Int, bigPow10(a.Commodity.Decimals))
	str := val.FloatString(p)
	if c == 'd' {
		str = addThousandSep(str, p)
	}
	if f.Flag(' ') && a.Sign() >= 0 {
		str = " " + str
	}
	format := "%"
	if f.Flag('-') {
		format += "-"
	}
	if w, ok := f.Width(); ok {
		fmt.Fprintf(f, format+"*s", w, str)
	} else {
		fmt.Fprintf(f, format+"s", str)
	}
}

func parseAmount(s string, c *Commodity) (*Amount, error) {
	ss := strings.Split(s, ".")
	ss[0] = strings.ReplaceAll(ss[0], ",", "")
	if len(ss) == 1 {
		i, err := strconv.ParseInt(ss[0], 10, 64)
		if err != nil {
			return nil, err
		}
		bi := big.NewInt(i)
		if c.Decimals > 0 {
			bi.Mul(bi, bigPow10(c.Decimals))
		}
		return NewAmount(bi, c), nil
	}
	if len(ss) != 2 {
		return nil, fmt.Errorf("malformed value %s", s)
	}
	s = strings.Join(ss, "")
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}
	bi := big.NewInt(i)
	decimals := len(ss[1])
	if c.Decimals != decimals {
		bi.Mul(bi, bigPow10(c.Decimals))
		bi.Div(bi, bigPow10(decimals))
	}
	return NewAmount(bi, c), nil
}

func (a *Amount) Write(w io.Writer, ledger bool) error {
	_, err := fmt.Fprintf(w, "%.*f %s", a.Decimals, a, a.SafeId(ledger))
	return err
}

func (b *Amount) adjustedTo(a *Amount) *big.Int {
	c := new(big.Int).Set(b.Int)
	if b.Decimals < a.Decimals {
		c.Mul(c, bigPow10(a.Decimals-b.Decimals))
	} else if a.Decimals < b.Decimals {
		c.Quo(c, bigPow10(b.Decimals-a.Decimals))
	}
	return c
}

func (a *Amount) AddIn(b *Amount) (err error) {
	if a.Commodity != b.Commodity {
		b, err = a.Commodity.Convert(b, b.Commodity)
		if err != nil {
			return err
		}
	}
	a.Add(a.Int, b.adjustedTo(a))
	return nil
}

func (a *Amount) Times(b *Amount) *Amount {
	c := new(big.Int).Mul(a.Int, b.Int)
	return &Amount{
		c.Quo(c, bigPow10(b.Decimals)),
		a.Commodity,
	}
}

func (a *Amount) IsZero() bool {
	return a == nil || a.Sign() == 0
}

func (a *Amount) IsEqual(b *Amount) bool    { return a.Cmp(b) == 0 }
func (a *Amount) IsLessThan(b *Amount) bool { return a.Cmp(b) < 0 }
func (a *Amount) IsMoreThan(b *Amount) bool { return a.Cmp(b) > 0 }
func (a *Amount) Cmp(b *Amount) int {
	return a.Int.Cmp(b.adjustedTo(a))
}

func (a *Amount) Magnitude() *big.Int {
	return new(big.Int).Abs(a.Int)
}

func (a *Amount) IsBigger(b *Amount) bool  { return a.CmpMagnitude(b) > 0 }
func (a *Amount) IsSmaller(b *Amount) bool { return a.CmpMagnitude(b) < 0 }
func (a *Amount) CmpMagnitude(b *Amount) int {
	return a.Magnitude().Cmp(b.Magnitude())
}

// func (a *Amount) DivBy(b *Amount) *Amount {
// 	c := b.adjustedTo(a)
// 	return &Amount{
// 		Int:      c.Div(a.Int, c),
// 		decimals: a.Commodity,
// 	}
// }

func (a *Amount) Negated() *Amount {
	return &Amount{
		new(big.Int).Neg(a.Int),
		a.Commodity,
	}
}

func MustParseAmount(f string, c *Commodity) *Amount {
	amt, err := parseAmount(f, c)
	if err != nil {
		panic(err)
	}
	return amt
}

func addThousandSep(str string, decimals int) string {
	start := 0
	if str[0] == '-' {
		start = 1
	}
	end := len(str)
	if decimals > 0 {
		end -= decimals + 1
	}
	n := end - start
	if n <= 3 {
		return str
	}
	var b strings.Builder
	b.Grow(len(str) + (n-1)/3)
	r := (n-1)%3 + 1
	b.WriteString(str[:start+r])
	for i := start + r; i < end; i += 3 {
		b.WriteByte(',')
		b.WriteString(str[i : i+3])
	}
	b.WriteString(str[end:])
	return b.String()
}

func (a *Amount) Width(decimals int) int {
	return a.width(decimals, false)
}

func (a *Amount) WidthWithDelimiters(decimals int) int {
	return a.width(decimals, true)
}

func (a *Amount) width(decimals int, thousandSep bool) int {
	w := bigLog10(a.Int) + 1
	if w <= a.Decimals {
		w = a.Decimals + 1
	}
	w = w - a.Decimals
	if thousandSep {
		w += (w - 1) / 3
	}
	w += decimals
	if decimals > 0 {
		w++ // decimal point
	}
	if a.Int.Sign() < 0 {
		w++ // minus sign
	}
	return w
}

func (a *Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%.*f %s", a.Commodity.Decimals, a, a.Commodity.Id))
}
