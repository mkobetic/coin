package main

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/mkobetic/coin"
)

func init() {
	(&cmdTags{}).newCommand("tags")
}

type cmdTags struct {
	flagsWithUsage
	values bool
}

func (*cmdTags) newCommand(names ...string) command {
	var cmd cmdTags
	cmd.FlagSet = newCommand(&cmd, names...)
	setUsage(cmd.FlagSet, `tags [flags] [NAMEREX]

List tags matching the NAMEREX.`)
	cmd.BoolVar(&cmd.values, "v", false, "print tag values if applicable")
	return &cmd
}

func (cmd *cmdTags) init() {
	coin.LoadAll()
}

func (cmd *cmdTags) execute(f io.Writer) {
	var nrex *regexp.Regexp
	if cmd.NArg() > 0 {
		nrex = regexp.MustCompile(cmd.Arg(0))
	}
	results := make(map[string][]string)
	for _, t := range coin.Transactions {
		collectKeys(nrex, t.Tags, results)
		for _, p := range t.Postings {
			collectKeys(nrex, p.Tags, results)
		}
	}
	for _, k := range sortAndClean(results) {
		vs := strings.Join(results[k], `", "`)
		colon := ""
		if len(vs) > 0 {
			colon = ":"
		}
		if cmd.values && len(vs) > 0 {
			fmt.Fprintf(f, `%s%s "%s"`+"\n", k, colon, vs)
		} else {
			fmt.Fprintf(f, "%s%s\n", k, colon)
		}
	}
}

func sortAndClean(results map[string][]string) (keys []string) {
	for k := range results {
		keys = append(keys, k)
		results[k] = clean(results[k])
	}
	sort.Strings(keys)
	return keys
}

func clean(list []string) []string {
	sort.Strings(list)
	if len(list[0]) == 0 {
		return list[1:]
	}
	return list
}

func collectKeys(nrex *regexp.Regexp, tags coin.Tags, results map[string][]string) {
	for k, v := range tags {
		if nrex == nil || nrex.MatchString(k) {
			results[k] = add(results[k], v)
		}
	}
}

func add(list []string, key string) []string {
	for _, v := range list {
		if v == key {
			return list
		}
	}
	return append(list, key)
}
