package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

func init() {
	(&cmdBalance{}).newCommand("balance", "bal", "b")
}

type cmdBalance struct {
	*flag.FlagSet
	begin, end  coin.Date
	zeroBalance bool
	level       int
}

func (*cmdBalance) newCommand(names ...string) command {
	var cmd cmdBalance
	cmd.FlagSet = newCommand(&cmd, names...)
	cmd.Var(&cmd.begin, "b", "begin balance from this date")
	cmd.Var(&cmd.end, "e", "end balance on this date")
	cmd.BoolVar(&cmd.zeroBalance, "z", false, "list accounts with zero total balance")
	cmd.IntVar(&cmd.level, "l", 0, "print accounts up to this level, 0 means all")
	return &cmd
}

func (cmd *cmdBalance) init() {
	coin.LoadAll()
}

func (cmd *cmdBalance) execute(f io.Writer) {
	account := coin.Root
	if cmd.NArg() > 0 {
		account = coin.MustFindAccount(cmd.Arg(0))
	}
	totals := make(balances)
	cumulative := make(balances)
	account.WithChildrenDo(func(a *coin.Account) {
		total := coin.NewZeroAmount(a.Commodity)
		for _, p := range trim(a.Postings, cmd.begin, cmd.end) {
			err := total.AddIn(p.Quantity)
			check.NoError(err, "adding posting for %s: %s\n", a.FullName, p.Transaction.Location())
		}
		totals[a] = total
		cumulative[a] = total.Copy()
	})
	account.FirstWithChildrenDo(func(a *coin.Account) {
		cump := cumulative[a.Parent]
		if cump == nil {
			return
		}
		cum := cumulative[a]
		err := cump.AddIn(cum)
		check.NoError(err, "cannot add total to parent of %s\n", a.FullName)
	})
	cmd.print(f, account, totals, cumulative)
}

func (cmd *cmdBalance) print(f io.Writer, acc *coin.Account, totals, cumulative balances) {
	width, cumWidth, curWidth := totals.maxWidth(), cumulative.maxWidth(), cumulative.curWidth()
	acc.WithChildrenDo(func(a *coin.Account) {
		if cmd.level != 0 && a.Depth() > cmd.level {
			return
		}
		if tot, cum := totals[a], cumulative[a]; cmd.zeroBalance || !cum.IsZero() {
			fmt.Fprintf(f, "%*a | %*a %-*s | %s\n",
				width, tot, cumWidth, cum, curWidth, a.CommodityId, a.FullName)
		}
	})
}

type balances map[*coin.Account]*coin.Amount

func (bs balances) maxWidth() int {
	var max int
	for acc, amt := range bs {
		if w := amt.Width(acc.Commodity.Decimals); w > max {
			max = w
		}
	}
	return max
}

// curWidth returns maximum currency width in the totals
func (bs balances) curWidth() int {
	var max int
	for _, b := range bs {
		if w := len(b.Commodity.Id); w > max {
			max = w
		}
	}
	return max
}
