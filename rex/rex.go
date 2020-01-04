package rex

import (
	"fmt"
	"regexp"
	"strconv"
)

type Exp struct {
	*regexp.Regexp
}

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
