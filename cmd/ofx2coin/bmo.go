package main

/*
	Bank of Montreal (BMO) produces an invalid ofx file when downloading
	multiple accounts at once.
	See https://github.com/aclindsa/ofxgo/issues/15 for more details.

	The bmoReader implements a hack that filters the invalid bits out, thus
	allowing to process these file correctly.
*/

import (
	"bufio"
	"bytes"
	"io"
)

var ignore = [][]byte{
	[]byte("<BANKMSGSET><BANKMSGSETV1>\n"),
	[]byte("<BANKMSGSETV1><BANKMSGSET>\n"),
}
var emptyLine = []byte("\n")

func newBMOReader(r io.Reader) *bmoReader {
	return &bmoReader{Reader: bufio.NewReader(r)}
}

type bmoReader struct {
	*bufio.Reader
	lastLine             []byte
	reachedHeader        bool // to be able to skip initial empty lines
	reachedBody          bool // to be able to inject missing empty line after header
	emptyLineAfterHeader bool
	// observed int
}

func (r *bmoReader) Read(p []byte) (n int, err error) {
	if len(r.lastLine) == 0 {
	loop:
		for {
			r.lastLine, err = r.ReadBytes('\n')
			if err != nil {
				return 0, err
			}
			if !r.reachedHeader {
				if bytes.Equal(r.lastLine, emptyLine) {
					continue loop
				} else {
					r.reachedHeader = true
				}
			}
			if !r.reachedBody {
				if bytes.Equal(r.lastLine, []byte("<OFX>\n")) {
					r.reachedBody = true
					if !r.emptyLineAfterHeader {
						n = copy(p, emptyLine)
						return n, nil
					}
				} else if bytes.Equal(r.lastLine, emptyLine) {
					r.emptyLineAfterHeader = true
				}
				break loop
			}
			for _, ignore := range ignore {
				if bytes.Equal(r.lastLine, ignore) {
					continue loop
				}
			}
			break loop
		}
	}
	n = copy(p, r.lastLine)
	r.lastLine = r.lastLine[n:]
	return n, nil
}
