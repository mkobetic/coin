// Package rex is a thin wrapper around the standard regexp package that allows interrogating a match result
// using the subexpression names rather than indices.
package rex

import (
	"fmt"
	"regexp"
	"strconv"
)

type Exp struct {
	*regexp.Regexp
}

// MustCompile returns a compiled regular expression with overriden Match function.
// MustCompile allows injecting other expressions into its argument using printf style formatting (%s),
// this is useful for composing expressions from other expressions.
func MustCompile(exp string, subexps ...*Exp) *Exp {
	if len(subexps) > 0 {
		var params []interface{}
		for _, s := range subexps {
			params = append(params, s)
		}
		exp = fmt.Sprintf(exp, params...)
	}
	rex := regexp.MustCompile(exp)
	return &Exp{rex}
}

// Match result is a map that maps subexpression names and indices to the corresponding values.
// Consequently the same subexpression can have multiple entries in the map, keyed by its index
// and by its name.
// Since subexpression names could be used multiple times in given expression, their occurrences
// have index suffix indicating the order in which they appeared. So if name X is used 3 times,
// there will entries X1, X2 and X3 in the result.
func (rex *Exp) Match(in []byte) (match map[string]string) {
	res := rex.FindSubmatch(in)
	if res == nil {
		return nil
	}
	byName := map[string][]string{}
	match = map[string]string{}
	for i, n := range rex.SubexpNames() {
		val := string(res[i])
		match[strconv.Itoa(i)] = val
		if n != "" {
			byName[n] = append(byName[n], val)
		}
	}
	for n, vals := range byName {
		if len(vals) == 1 {
			match[n] = vals[0]
			continue
		}
		for i, val := range vals {
			match[n+strconv.Itoa(i+1)] = val
		}
	}
	return match
}
