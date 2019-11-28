package warn

import (
	"fmt"
	"os"
)

func If(holds bool, format string, args ...interface{}) {
	if !holds {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}
