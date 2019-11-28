package coin

import (
	"bytes"
	"fmt"
	"regexp"
)

type Test struct {
	Cmd    []byte
	Result []byte
}

var testHead = regexp.MustCompile(`test\s+(.+)`)

func (p *Parser) parseTest() (*Test, error) {
	matches := testHead.FindSubmatch(p.Bytes())
	t := &Test{Cmd: matches[1]}
	var b bytes.Buffer
	for p.Scan() {
		if bytes.Equal((bytes.TrimSpace(p.Bytes())), []byte("end test")) {
			t.Result = b.Bytes()
			return t, nil
		}
		b.Write(p.Bytes())
		fmt.Fprintln(&b)
	}
	return t, p.Err()
}
