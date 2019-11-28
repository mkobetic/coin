package check

import (
	"fmt"
	"os"
)

func If(holds bool, format string, args ...interface{}) {
	if holds {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

var OK = If

func NoError(err error, format string, args ...interface{}) {
	If(err == nil, "Error: %s\n"+format, append([]interface{}{err}, args...)...)
}

func Equal(a, b interface{}, format string, args ...interface{}) {
	If(a == b, "Not equal\n%v\n%v\n"+format, append([]interface{}{a, b}, args...)...)
}
