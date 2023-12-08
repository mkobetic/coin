package coin

import (
	"fmt"
	"testing"

	"github.com/mkobetic/coin/assert"
)

func Test_ParseTags(t *testing.T) {
	for i, test := range []struct {
		line string
		tags string
	}{
		{
			line: "hello #foo hi",
			tags: `map[foo:]`,
		},
		{
			line: "hello #foo:   hi ho   ",
			tags: `map[foo:hi ho]`,
		},
		{
			line: "#foo: hi, #bar there #baz: and now what",
			tags: `map[bar: baz:and now what foo:hi]`,
		},
		{
			line: "hello #foo: hi, #bar: there, #baz: now",
			tags: `map[bar:there baz:now foo:hi]`,
		},
	} {
		t.Run(fmt.Sprintf("%d %s", i, test.line), func(t *testing.T) {
			tags := ParseTags(test.line)
			out := fmt.Sprintf("%v", tags)
			assert.Equal(t, test.tags, out)
		})
	}
}
