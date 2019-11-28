package coin

import (
	"math/big"
	"sort"
)

var (
	log10s    []int64
	bigPow10s []*big.Int
	big1      = big.NewInt(1)
	big10     = big.NewInt(10)
	big100    = big.NewInt(100)
	big1000   = big.NewInt(1000)
	big10000  = big.NewInt(10000)
)

func init() {
	log10s = make([]int64, 18)
	x := int64(10)
	for i := range log10s {
		log10s[i] = x
		x = x * 10
	}
	bigPow10s = make([]*big.Int, 50)
	prev := big1
	for i := range bigPow10s {
		next := new(big.Int)
		next.Mul(prev, big10)
		bigPow10s[i] = next
		prev = next
	}
}

func log10(i int64) int {
	switch {
	case i < 10:
		return 0
	case i < 100:
		return 1
	case i < 1000:
		return 2
	case i < 10000:
		return 3
	}
	return sort.Search(
		len(log10s),
		func(j int) bool { return i < log10s[j] },
	)
}

func pow10(i int) int64 {
	if i < 1 {
		return 1
	}
	return log10s[i-1]
}

func bigLog10(bi *big.Int) int {
	for i, bp10 := range bigPow10s {
		if bi.CmpAbs(bp10) < 0 {
			return i
		}
	}
	panic("number too big")
}

func bigPow10(i int) *big.Int {
	if i < 1 {
		return big1
	}
	return bigPow10s[i-1]
}
