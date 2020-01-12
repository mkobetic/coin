package coin

import "testing"

import "github.com/mkobetic/coin/assert"

func Test_ParseDate(t *testing.T) {
	Year, Month, Day = 2019, 10, 22
	for in, out := range map[string]string{
		"2012/12/12": "2012/12/12",
		"95/12/2":    "1995/12/02",
		"66/7/12":    "2066/07/12",
		"3/12":       "2020/03/12",
		"06/12":      "2019/06/12",
		"2020/06":    "2020/06/01",
		"2020":       "2020/01/01",
		"+3d":        "2019/10/25",
		"-2w":        "2019/10/08",
		"+46m":       "2023/08/22",
		"-350y":      "1669/10/22",
		"1/1+6w":     "2020/02/12",
	} {
		match := DateREX.Match([]byte(in))
		assert.NotNil(t, match)
		dt, err := parseDate(match, 0)
		assert.NoError(t, err)
		got := dt.Format(DateFormat)
		assert.Equal(t, got, out)
	}
}

func Test_Date(t *testing.T) {
	var d Date
	assert.NoError(t, (&d).Set("2012/12/12"))
	assert.Equal(t, d.String(), "2012/12/12")
}
