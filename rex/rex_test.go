package rex

import (
	"fmt"
	"github.com/mkobetic/coin/assert"
	"testing"
)

func Test_All(t *testing.T) {
	date := MustCompile(`(?P<date>((?P<year>\d{4})/)?(?P<month>\d\d)/(?P<day>\d\d))`)
	offset := MustCompile(`(?P<offset>[+-]\d+[d|w|m|y])`)
	commodity := MustCompile(`(?P<commodity>\w+)`)
	amount := MustCompile(`(?P<amount>(?P<quantity>\d+(\.\d+)?)\s+%s)`, commodity)
	posted := MustCompile(`(?P<posted>%s?%s?)`, date, offset)
	rex := MustCompile(`^%s\s+%s(\s+@\s+%s)?`, posted, amount, amount)
	for in, out := range map[string]string{
		"12/11 15.22 CAD": `` +
			`map[0:12/11 15.22 CAD 1:12/11 10:.22 11:CAD 12: 13: 14: 15: 16: 2:12/11 3: 4: 5:12 6:11 7: 8:15.22 CAD 9:15.22 ` +
			`amount1:15.22 CAD amount2: commodity1:CAD commodity2: date:12/11 day:11 month:12 offset: posted:12/11 quantity1:15.22 quantity2: year:]`,
		"2020/12/11-5y 100 books": `` +
			`map[0:2020/12/11-5y 100 books 1:2020/12/11-5y 10: 11:books 12: 13: 14: 15: 16: 2:2020/12/11 3:2020/ 4:2020 5:12 6:11 7:-5y 8:100 books 9:100 ` +
			`amount1:100 books amount2: commodity1:books commodity2: date:2020/12/11 day:11 month:12 offset:-5y posted:2020/12/11-5y quantity1:100 quantity2: year:2020]`,
		"-5m   183   VBAL   @ 5.55  USD": `` +
			`map[0:-5m   183   VBAL   @ 5.55  USD 1:-5m 10: 11:VBAL 12:   @ 5.55  USD 13:5.55  USD 14:5.55 15:.55 16:USD 2: 3: 4: 5: 6: 7:-5m 8:183   VBAL 9:183 ` +
			`amount1:183   VBAL amount2:5.55  USD commodity1:VBAL commodity2:USD date: day: month: offset:-5m posted:-5m quantity1:183 quantity2:5.55 year:]`,
	} {
		match := rex.Match([]byte(in))
		assert.Equal(t, fmt.Sprint(match), out)
	}
}
