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

func newBMOReader(r io.Reader) *bmoReader {
	return &bmoReader{Reader: bufio.NewReader(r)}
}

type bmoReader struct {
	*bufio.Reader
	lastLine []byte
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
