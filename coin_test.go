package coin

import (
	"testing"
)

func Test_MatchAccountName(t *testing.T) {
	for _, tf := range []struct {
		pattern string
		matches []string
		fails   []string
	}{
		{
			"ex",
			[]string{"Expenses", "Income:Extra", "Investments:Forex:USD"},
			[]string{"Fees", "Root"},
		},
		{
			"^ex",
			[]string{"Expenses"},
			[]string{"Income:Extra", "Investments:Forex:USD"},
		},
		{
			"e::c",
			[]string{"Expenses:Travel:CAD", "Expenses:CAD", "Investments:Forex:CAD"},
			[]string{"Expenses", "Root"},
		},
		{
			"l::mc$",
			[]string{"Liabilities:Credit:MC"},
			[]string{"Liabilities:Credit:MC2", "Liabilities:Credit:MC:XX"},
		},
	} {
		rx := ToRegex(tf.pattern)
		for _, m := range tf.matches {
			if !rx.MatchString(m) {
				t.Errorf("%s should match %s\nregexp: %s", tf.pattern, m, rx.String())
			}
		}
		for _, m := range tf.fails {
			if rx.MatchString(m) {
				t.Errorf("%s should not match %s", tf.pattern, m)
			}
		}
	}
}
