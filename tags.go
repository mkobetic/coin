package coin

import "regexp"

var tagREX = regexp.MustCompile(`#(?P<key>\w+)(:\s*(?P<value>[^,]+\S)\s*(,|$))?`)
var tagREXKey = tagREX.SubexpIndex("key")
var tagREXValue = tagREX.SubexpIndex("value")

type Tags map[string]string

func parseTags(lines []string) Tags {
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
