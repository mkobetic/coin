package main

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/mkobetic/coin/assert"
)

func Test_BMOReader(t *testing.T) {
	in := `
hello
<BANKMSGSET><BANKMSGSETV1>
world


and
<BANKMSGSET>
again
<BANKMSGSET><BANKMSGSETV1>
and
done
<BANKMSGSET><BANKMSGSETV1>
`
	out := `
hello
world


and
<BANKMSGSET>
again
and
done
`
	r := newBMOReader(strings.NewReader(in))
	got, err := ioutil.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, string(got), out)
}
