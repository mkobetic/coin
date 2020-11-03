package main

import (
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

// These are used for embedding the report files.
// see embed.go and reports.go

func decode(name, encoded string) (decoded []byte) {
	r, err := gzip.NewReader(base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded)))
	check.NoError(err, "opening %s", name)
	decoded, err = ioutil.ReadAll(r)
	check.NoError(err, "reading %s", name)
	return decoded
}

func trim(ps []*coin.Posting, begin, end coin.Date) []*coin.Posting {
	if !begin.IsZero() {
		from := sort.Search(len(ps), func(i int) bool {
			return !ps[i].Transaction.Posted.Before(begin.Time)
		})
		if from == len(ps) {
			return nil
		}
		ps = ps[from:]
	}
	if !end.IsZero() {
		to := sort.Search(len(ps), func(i int) bool {
			return !ps[i].Transaction.Posted.Before(end.Time)
		})
		if to == len(ps) {
			return ps
		}
		ps = ps[:to]
	}
	return ps
}
