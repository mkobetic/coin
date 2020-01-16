package coin

import (
	"bytes"
	"fmt"
	"regexp"
)

type Test struct {
	Cmd    []byte
	Result []byte

	line uint
	file string
}

var testHead = regexp.MustCompile(`test\s+(.+)`)

func (p *Parser) parseTest(fn string) (*Test, error) {
	matches := testHead.FindSubmatch(p.Bytes())
	t := &Test{Cmd: matches[1], line: p.lineNr, file: fn}
	var b bytes.Buffer
	for p.Scan() {
		if bytes.Equal((bytes.TrimSpace(p.Bytes())), []byte("end test")) {
			t.Result = b.Bytes()
			p.Scan()
			return t, nil
		}
		b.Write(p.Bytes())
		fmt.Fprintln(&b)
	}
	return t, p.Err()
}

func (t *Test) Location() string {
	return fmt.Sprintf("%s:%d", t.file, t.line)
}
