package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

var cwd string // caches current working directory

func init() {
	cwd, _ = os.Getwd()
}

func trimLocation(loc string) string {
	l := strings.TrimPrefix(loc, cwd)
	if len(l) < len(loc) {
		return "." + l
	}
	return loc
}

type postings []*coin.Posting

func (ps postings) widths(acctPrefix string) (widths [4]int) {
	for _, p := range ps {
		widths[0] = max(widths[0], len(p.Transaction.Description))
		widths[1] = max(widths[1], len(strings.TrimPrefix(p.Account.FullName, acctPrefix)))
		widths[2] = max(widths[2], p.Quantity.Width(p.Account.Commodity.Decimals))
		widths[3] = max(widths[3], len(p.Transaction.Other(p).Account.FullName))
	}
	return widths
}

func (ps postings) totals(com *coin.Commodity) (ts []*coin.Amount) {
	total := coin.NewZeroAmount(com)
	for _, p := range ps {
		err := total.AddIn(p.Quantity)
		check.NoError(err, "adding posting for %s: %s\n", p.Account.FullName, p.Transaction.Location())
		ts = append(ts, total.Copy())
	}
	return ts
}

func (ps postings) print(f io.Writer, opts *options) {
	if len(ps) == 0 {
		return
	}
	widths := ps.widths(opts.Prefix())
	widths[0] = min(widths[0], opts.MaxDesc())
	widths[3] = min(widths[3], opts.MaxAcct())
	commodity := opts.commodity
	if commodity == nil {
		commodity = ps[0].Account.Commodity
	}
	totals := ps.totals(commodity)
	tWidth := totals[len(totals)-1].Width(commodity.Decimals)
	fmtString := "%s | %*s | %*s | %*a | %*a %s%c\n"
	if opts.Location() {
		fmtString = "%s | %*s | %*s | %*a | %*a %s%c| %s\n"
	}
	for i, s := range ps {
		reconciled := ' '
		if s.BalanceAsserted {
			reconciled = '*'
		}
		args := []interface{}{
			s.Transaction.Posted.Format(coin.DateFormat),
			widths[0], s.Transaction.Description,
			widths[3], coin.ShortenAccountName(strings.TrimPrefix(s.Transaction.Other(s).Account.FullName, opts.Prefix()), opts.MaxAcct()),
			widths[2], s.Quantity,
			tWidth, totals[i],
			s.Account.CommodityId,
			reconciled,
		}
		if opts.Location() {
			args = append(args, trimLocation(s.Transaction.Location()))
		}
		fmt.Fprintf(f, fmtString, args...)
		if opts.showNotes && (len(s.Notes) > 0 || len(s.Transaction.Notes) > 0) {
			printNotes(f, strings.Repeat(" ", len(args[0].(string)))+" ;", s)
		}
	}
}

func (ps postings) printLong(f io.Writer, opts *options) {
	if len(ps) == 0 {
		return
	}
	widths := ps.widths(opts.Prefix())
	widths[0] = min(widths[0], opts.MaxDesc())
	widths[1] = min(widths[1], opts.MaxAcct())
	widths[3] = min(widths[3], opts.MaxAcct())
	commodity := opts.commodity
	if commodity == nil {
		commodity = ps[0].Account.Commodity
	}
	totals := ps.totals(commodity)
	tWidth := totals[len(totals)-1].Width(commodity.Decimals)
	fmtString := "%s | %*s | %*s | %*s | %*a | %*a %s%c\n"
	if opts.Location() {
		fmtString = "%s | %*s | %*s | %*s | %*a | %*a %s%c| %s\n"
	}
	for i, s := range ps {
		reconciled := ' '
		if s.BalanceAsserted {
			reconciled = '*'
		}
		args := []interface{}{
			s.Transaction.Posted.Format(coin.DateFormat),
			widths[0], s.Transaction.Description,
			widths[1], coin.ShortenAccountName(strings.TrimPrefix(s.Account.FullName, opts.Prefix()), opts.MaxAcct()),
			widths[3], coin.ShortenAccountName(strings.TrimPrefix(s.Transaction.Other(s).Account.FullName, opts.Prefix()), opts.MaxAcct()),
			widths[2], s.Quantity,
			tWidth, totals[i],
			s.Account.CommodityId,
			reconciled,
		}
		if opts.Location() {
			args = append(args, trimLocation(s.Transaction.Location()))
		}
		fmt.Fprintf(f, fmtString, args...)
		if opts.showNotes && (len(s.Notes) > 0 || len(s.Transaction.Notes) > 0) {
			printNotes(f, strings.Repeat(" ", len(args[0].(string)))+" ;", s)
		}
	}
}

func printNotes(w io.Writer, prefix string, p *coin.Posting) {
	for _, line := range append(p.Notes, p.Transaction.Notes...) {
		fmt.Fprintln(w, prefix, line)
	}
}

type options struct {
	prefix           string
	location         bool
	maxDesc, maxAcct int
	commodity        *coin.Commodity
	showNotes        bool
}

func (o *options) MaxDesc() int {
	if o == nil || o.maxDesc == 0 {
		return 50
	}
	return o.maxDesc
}

func (o *options) MaxAcct() int {
	if o == nil || o.maxAcct == 0 {
		return 15
	}
	return o.maxAcct
}

func (o *options) Prefix() string {
	if o == nil {
		return ""
	}
	return o.prefix
}

func (o *options) Location() bool {
	if o == nil {
		return false
	}
	return o.location
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
