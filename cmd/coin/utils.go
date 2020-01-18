package main

import (
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"strings"

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
