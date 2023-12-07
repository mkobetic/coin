package main

import (
	"bufio"
	"strings"
)

func trim(in string) string {
	scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(in)))
	scanner.Split(bufio.ScanWords)
	var w strings.Builder
	isFirst := true
	for scanner.Scan() {
		if !isFirst {
			w.WriteByte(' ')
		}
		w.Write(scanner.Bytes())
		isFirst = false
	}
	return w.String()
}
