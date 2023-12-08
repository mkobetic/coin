package coin

import (
	"regexp"
	"sort"
)

var tagREX = regexp.MustCompile(`#(?P<key>\w+)(:\s*(?P<value>[^,]+\S)\s*(,|$))?`)
var tagREXKey = tagREX.SubexpIndex("key")
var tagREXValue = tagREX.SubexpIndex("value")

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

func (t Tags) Has(rex *regexp.Regexp) bool {
	if t == nil {
		return false
	}
	for k := range t {
		if rex.MatchString(k) {
			return true
		}
	}
	return false
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
