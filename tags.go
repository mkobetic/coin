package coin

import (
	"regexp"
	"sort"
	"strings"
)

var tagREX = regexp.MustCompile(`#(?P<key>[\w-/]+)(:\s*(?P<value>[^,]+\S)\s*(,|$))?`)
var tagREXKey = tagREX.SubexpIndex("key")
var tagREXValue = tagREX.SubexpIndex("value")

// Tags represents tags associated with a posting or transaction.
// A nil value is also valid.
type Tags map[string]string

func ParseTags(lines ...string) Tags {
	tags := make(Tags)
	for _, line := range lines {
		for _, match := range tagREX.FindAllStringSubmatch(line, -1) {
			key, value := match[tagREXKey], match[tagREXValue]
			tags[key] = value
		}
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}

func (t Tags) Includes(key string) bool {
	if t == nil {
		return false
	}
	_, ok := t[key]
	return ok
}

func (t Tags) Value(key string) string {
	if t == nil {
		return ""
	}
	return t[key]
}

func (t Tags) Keys() (keys []string) {
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (t Tags) KeysAndValues() (kvs []string) {
	for k, v := range t {
		if v == "" {
			kvs = append(kvs, k)
		} else {
			kvs = append(kvs, k+":"+v)
		}
	}
	sort.Strings(kvs)
	return kvs
}

// TagMatcher matches a posting or transaction against a tag expression.
// Tag expression is one or two regular expressions separated by a colon,
// matched against a tag key and optionally a tag value.
type TagMatcher struct {
	Key, Value *regexp.Regexp
	inverted   bool // match if key/value not matching
}

func NewTagMatcher(exp string) *TagMatcher {
	if len(exp) == 0 {
		return nil
	}
	inverted := false
	if exp[0] == '!' {
		inverted = true
		exp = exp[1:]
	}
	parts := strings.SplitN(exp, ":", 2)
	var key, value *regexp.Regexp
	key = regexp.MustCompile("(?i)" + parts[0])
	if len(parts) > 1 {
		value = regexp.MustCompile("(?i)" + parts[1])
	}
	return &TagMatcher{Key: key, Value: value, inverted: inverted}
}

func (m *TagMatcher) Match(tags Tags) bool {
	if tags == nil {
		return false
	}
	for k, v := range tags {
		if match := m.Key.MatchString(k) && (m.Value == nil || m.Value.MatchString(v)); match && !m.inverted || !match && m.inverted {
			return true
		}
	}
	return false
}
