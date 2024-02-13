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
	fValues   bool
	fAccounts bool

	results  map[string][]string
	accounts map[string][]string
}

func (*cmdTags) newCommand(names ...string) command {
	var cmd cmdTags
	cmd.FlagSet = newCommand(&cmd, names...)
	setUsage(cmd.FlagSet, `tags [flags] [NAMEREX]

List tags matching the NAMEREX.`)
	cmd.BoolVar(&cmd.fValues, "v", false, "print tag values if applicable")
	cmd.BoolVar(&cmd.fAccounts, "a", false, "print account names where tag is used")
	return &cmd
}

func (cmd *cmdTags) init() {
	coin.LoadAll()
}

func (cmd *cmdTags) execute(f io.Writer) {
	var nrex *regexp.Regexp
	if cmd.NArg() > 0 {
		nrex = regexp.MustCompile("(?i)" + cmd.Arg(0))
	}
	cmd.results = make(map[string][]string)
	cmd.accounts = make(map[string][]string)
	for _, t := range coin.Transactions {
		accounts := [](*coin.Account){}
		for _, p := range t.Postings {
			cmd.collectKeys(nrex, p.Tags, p.Account)
			if cmd.fAccounts {
				accounts = append(accounts, p.Account)
			}
		}
		cmd.collectKeys(nrex, t.Tags, accounts...)
	}
	for _, k := range sortAndClean(cmd.results) {
		vs := strings.Join(cmd.results[k], `", "`)
		colon := ""
		if len(vs) > 0 {
			colon = ":"
		}
		if cmd.fValues && len(vs) > 0 {
			fmt.Fprintf(f, `%s%s "%s"`+"\n", k, colon, vs)
		} else {
			fmt.Fprintf(f, "%s%s\n", k, colon)
		}
		if cmd.fAccounts {
			for _, a := range cmd.accounts[k] {
				fmt.Fprintf(f, "\t%s\n", a)
			}
		}
	}
}

func (cmd *cmdTags) collectKeys(nrex *regexp.Regexp, tags coin.Tags, accounts ...*coin.Account) {
	for k, v := range tags {
		if nrex == nil || nrex.MatchString(k) {
			cmd.results[k] = add(cmd.results[k], v)
			if cmd.fAccounts {
				list := cmd.accounts[k]
				for _, a := range accounts {
					list = add(list, a.FullName)
				}
				cmd.accounts[k] = list
			}
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

func add(list []string, key string) []string {
	for _, v := range list {
		if v == key {
			return list
		}
	}
	return append(list, key)
}
