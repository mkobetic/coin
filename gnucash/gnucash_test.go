package gnucash

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/assert"
)

func Test_Amount(t *testing.T) {
	cad := &coin.Commodity{Id: "CAD", Decimals: 2}
	for i, fix := range []struct {
		in  string
		out string
	}{
		{"-1000/1", "-1000.00"},
		{"-100000/100", "-1000.00"},
		{"-100000/200", "-500.00"},
		{"-100000/2000", "-50.00"},
		{"-100000/20000", "-5.00"},
	} {
		amt := mustParseAmount(fix.in, cad)
		res := fmt.Sprintf("%a", amt)
		assert.Equal(t, res, fix.out, "%d. not equal", i)
	}
}

func Test_Unmarshaling(t *testing.T) {
	db := Gnucash{Book: Book{}}
	err := xml.Unmarshal(Sample, &db)
	if err != nil {
		t.Error(err)
	}
	db.Book.Resolve()

	exp := []string{
		"           0.00          237.04 CAD        Root [0]",
		"       -9765.09          154.53 CAD        Assets [1]",
		"        7729.58         7729.58 CAD        Assets:Bank [4]",
		"         82.612          82.612 ZLB        Assets:Investments [1]",
		"           0.00           82.51 CAD        Expenses [0]",
		"          82.51           82.51 CAD        Expenses:Fuel [2]",
	}
	i := 0
	coin.Root.WithChildrenDo(func(a *coin.Account) {
		assert.Equal(t, a.String(), exp[i])
		i++
	})
	assert.Equal(t, len(exp), i)

	for id, exp := range map[string]string{
		"CAD": "CAD: 2013/11/13 0.94 USD [2]",
		"USD": "USD",
		"ZLB": "ZLB: 2015/12/08 26.51 CAD [1]",
	} {
		assert.Equal(t, coin.Commodities[id].String(), exp)
	}
	for i, exp := range []string{
		"P 2013/11/05 CAD 0.94 USD\n",
		"P 2013/11/13 CAD 0.94 USD\n",
		"P 2015/12/08 ZLB 26.51 CAD\n",
	} {
		assert.Equal(t, coin.Prices[i].String(), exp)
	}
	for i, exp := range []string{
		"2007/12/10 Contribution\n" +
			"  Assets:Bank   9765.09 CAD\n" +
			"  Assets       -9765.09 CAD\n",
		"2007/12/12 Investment\n" +
			"  Assets:Investments    82.612 ZLB\n" +
			"  Assets:Bank         -1953.00 CAD\n",
		"2011/09/15 Sienna\n" +
			"  Expenses:Fuel   45.00 CAD\n" +
			"  Assets:Bank    -45.00 CAD\n",
		"2011/09/18 Sienna\n" +
			"  Expenses:Fuel   37.51 CAD\n" +
			"  Assets:Bank    -37.51 CAD\n",
	} {
		assert.Equal(t, coin.Transactions[i].String(), exp)
	}
}

func Test_KvpValue(t *testing.T) {
	sample := []byte(`
<slot>
  <key>mykey</key>
  <value type="mytype">myvalue</value>
</slot>`)
	var v KvpSlot
	err := xml.Unmarshal(sample, &v)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, v.Key, "mykey")
	assert.Equal(t, v.Value.Type, "mytype")
	assert.Equal(t, v.Value.Value, "myvalue")
}

var Sample = []byte(`
<?xml version="1.0" encoding="utf-8" ?>
<gnc-v2
     xmlns:gnc="http://www.gnucash.org/XML/gnc"
     xmlns:act="http://www.gnucash.org/XML/act"
     xmlns:book="http://www.gnucash.org/XML/book"
     xmlns:cd="http://www.gnucash.org/XML/cd"
     xmlns:cmdty="http://www.gnucash.org/XML/cmdty"
     xmlns:price="http://www.gnucash.org/XML/price"
     xmlns:slot="http://www.gnucash.org/XML/slot"
     xmlns:split="http://www.gnucash.org/XML/split"
     xmlns:sx="http://www.gnucash.org/XML/sx"
     xmlns:trn="http://www.gnucash.org/XML/trn"
     xmlns:ts="http://www.gnucash.org/XML/ts"
     xmlns:fs="http://www.gnucash.org/XML/fs"
     xmlns:bgt="http://www.gnucash.org/XML/bgt"
     xmlns:recurrence="http://www.gnucash.org/XML/recurrence"
     xmlns:lot="http://www.gnucash.org/XML/lot"
     xmlns:addr="http://www.gnucash.org/XML/addr"
     xmlns:owner="http://www.gnucash.org/XML/owner"
     xmlns:billterm="http://www.gnucash.org/XML/billterm"
     xmlns:bt-days="http://www.gnucash.org/XML/bt-days"
     xmlns:bt-prox="http://www.gnucash.org/XML/bt-prox"
     xmlns:cust="http://www.gnucash.org/XML/cust"
     xmlns:employee="http://www.gnucash.org/XML/employee"
     xmlns:entry="http://www.gnucash.org/XML/entry"
     xmlns:invoice="http://www.gnucash.org/XML/invoice"
     xmlns:job="http://www.gnucash.org/XML/job"
     xmlns:order="http://www.gnucash.org/XML/order"
     xmlns:taxtable="http://www.gnucash.org/XML/taxtable"
     xmlns:tte="http://www.gnucash.org/XML/tte"
     xmlns:vendor="http://www.gnucash.org/XML/vendor">
<gnc:count-data cd:type="book">1</gnc:count-data>
<gnc:book version="2.0.0">
<book:id type="guid">3dc1c84173022dfdb9410e0933fa2fa3</book:id>
<book:slots>
  <slot>
    <slot:key>options</slot:key>
    <slot:value type="frame">
      <slot>
        <slot:key>Budgeting</slot:key>
        <slot:value type="frame"/>
      </slot>
    </slot:value>
  </slot>
</book:slots>
<gnc:count-data cd:type="commodity">3</gnc:count-data>
<gnc:count-data cd:type="account">3</gnc:count-data>
<gnc:count-data cd:type="transaction">4</gnc:count-data>
<gnc:count-data cd:type="price">3</gnc:count-data>
<gnc:commodity version="2.0.0">
  <cmdty:space>ISO4217</cmdty:space>
  <cmdty:id>CAD</cmdty:id>
  <cmdty:fraction>100</cmdty:fraction>
  <cmdty:get_quotes/>
  <cmdty:quote_source>currency</cmdty:quote_source>
  <cmdty:quote_tz/>
</gnc:commodity>
<gnc:commodity version="2.0.0">
  <cmdty:space>ISO4217</cmdty:space>
  <cmdty:id>USD</cmdty:id>
  <cmdty:fraction>100</cmdty:fraction>
  <cmdty:get_quotes/>
  <cmdty:quote_source>currency</cmdty:quote_source>
  <cmdty:quote_tz/>
</gnc:commodity>
<gnc:commodity version="2.0.0">
  <cmdty:space>TSX</cmdty:space>
  <cmdty:id>ZLB</cmdty:id>
  <cmdty:name>BMO Low Volatility Canadian Equity ETF</cmdty:name>
  <cmdty:xcode>05573T102</cmdty:xcode>
  <cmdty:fraction>1000</cmdty:fraction>
  <cmdty:get_quotes/>
  <cmdty:quote_source>yahoo</cmdty:quote_source>
  <cmdty:quote_tz/>
  <cmdty:slots>
    <slot>
      <slot:key>user_symbol</slot:key>
      <slot:value type="string">ZLB</slot:value>
    </slot>
  </cmdty:slots>
</gnc:commodity>
<gnc:pricedb version="1">
  <price>
    <price:id type="guid">3b036a168478f78e0cb62857a7d19467</price:id>
    <price:commodity>
      <cmdty:space>ISO4217</cmdty:space>
      <cmdty:id>CAD</cmdty:id>
    </price:commodity>
    <price:currency>
      <cmdty:space>ISO4217</cmdty:space>
      <cmdty:id>USD</cmdty:id>
    </price:currency>
    <price:time>
      <ts:date>2013-11-13 00:00:00 -0500</ts:date>
    </price:time>
    <price:source>user:xfer-dialog</price:source>
    <price:value>200/211</price:value>
  </price>
  <price>
    <price:id type="guid">12008009939d2cc55898df7f10a2e025</price:id>
    <price:commodity>
      <cmdty:space>ISO4217</cmdty:space>
      <cmdty:id>CAD</cmdty:id>
    </price:commodity>
    <price:currency>
      <cmdty:space>ISO4217</cmdty:space>
      <cmdty:id>USD</cmdty:id>
    </price:currency>
    <price:time>
      <ts:date>2013-11-05 00:00:00 -0500</ts:date>
    </price:time>
    <price:source>user:xfer-dialog</price:source>
    <price:value>1000/1061</price:value>
  </price>
  <price>
    <price:id type="guid">b7dec46e32f9131c90781e108051806a</price:id>
    <price:commodity>
      <cmdty:space>TSX</cmdty:space>
      <cmdty:id>ZLB</cmdty:id>
    </price:commodity>
    <price:currency>
      <cmdty:space>ISO4217</cmdty:space>
      <cmdty:id>CAD</cmdty:id>
    </price:currency>
    <price:time>
      <ts:date>2015-12-08 00:00:00 -0500</ts:date>
    </price:time>
    <price:source>user:xfer-dialog</price:source>
    <price:value>265199/10000</price:value>
  </price>
</gnc:pricedb>
<gnc:account version="2.0.0">
  <act:name>Root Account</act:name>
  <act:id type="guid">5109123737a9299489487a430a954de7</act:id>
  <act:type>ROOT</act:type>
</gnc:account>
<gnc:account version="2.0.0">
  <act:name>Assets</act:name>
  <act:id type="guid">6c736c815b25f0ab5b7b96571e15812b</act:id>
  <act:type>ASSET</act:type>
  <act:commodity>
    <cmdty:space>ISO4217</cmdty:space>
    <cmdty:id>CAD</cmdty:id>
  </act:commodity>
  <act:commodity-scu>100</act:commodity-scu>
  <act:description>Assets</act:description>
  <act:parent type="guid">5109123737a9299489487a430a954de7</act:parent>
</gnc:account>
<gnc:account version="2.0.0">
  <act:name>Bank</act:name>
  <act:id type="guid">53f6507e7492d6fb0600772d2da50cff</act:id>
  <act:type>ASSET</act:type>
  <act:commodity>
    <cmdty:space>ISO4217</cmdty:space>
    <cmdty:id>CAD</cmdty:id>
  </act:commodity>
  <act:commodity-scu>100</act:commodity-scu>
  <act:parent type="guid">6c736c815b25f0ab5b7b96571e15812b</act:parent>
</gnc:account>
<gnc:account version="2.0.0">
  <act:name>Investments</act:name>
  <act:id type="guid">b761bc5790b7d1018aef110c10921a45</act:id>
  <act:type>ASSET</act:type>
  <act:commodity>
    <cmdty:space>TSX</cmdty:space>
    <cmdty:id>ZLB</cmdty:id>
  </act:commodity>
  <act:commodity-scu>100</act:commodity-scu>
  <act:description>Groups Various Investment Accounts</act:description>
  <act:parent type="guid">6c736c815b25f0ab5b7b96571e15812b</act:parent>
</gnc:account>
<gnc:account version="2.0.0">
  <act:name>Expenses</act:name>
  <act:id type="guid">214a5b710177100fd2f652b6e9db8e92</act:id>
  <act:type>EXPENSE</act:type>
  <act:commodity>
    <cmdty:space>ISO4217</cmdty:space>
    <cmdty:id>CAD</cmdty:id>
  </act:commodity>
  <act:commodity-scu>100</act:commodity-scu>
  <act:description>Expenses</act:description>
  <act:parent type="guid">5109123737a9299489487a430a954de7</act:parent>
</gnc:account>
<gnc:account version="2.0.0">
  <act:name>Fuel</act:name>
  <act:id type="guid">fed7736c5ecc376fabd1a96d6aac8218</act:id>
  <act:type>EXPENSE</act:type>
  <act:commodity>
    <cmdty:space>ISO4217</cmdty:space>
    <cmdty:id>CAD</cmdty:id>
  </act:commodity>
  <act:commodity-scu>100</act:commodity-scu>
  <act:description>Fuel expenses</act:description>
  <act:parent type="guid">214a5b710177100fd2f652b6e9db8e92</act:parent>
</gnc:account>
<gnc:transaction version="2.0.0">
  <trn:id type="guid">dc41a515493c41147f4f48cf5a0ab6a7</trn:id>
  <trn:currency>
    <cmdty:space>ISO4217</cmdty:space>
    <cmdty:id>CAD</cmdty:id>
  </trn:currency>
  <trn:date-posted>
    <ts:date>2007-12-10 05:59:00 -0500</ts:date>
  </trn:date-posted>
  <trn:date-entered>
    <ts:date>2010-03-14 22:52:08 -0400</ts:date>
  </trn:date-entered>
  <trn:description>Contribution</trn:description>
  <trn:splits>
    <trn:split>
      <split:id type="guid">6ad38aa95fbc9b6cc015fa84bf97e33e</split:id>
      <split:reconciled-state>n</split:reconciled-state>
      <split:value>976509/100</split:value>
      <split:quantity>976509/100</split:quantity>
      <split:account type="guid">53f6507e7492d6fb0600772d2da50cff</split:account>
    </trn:split>
    <trn:split>
      <split:id type="guid">215f4f3f131c44cf9b796f50887bc51a</split:id>
      <split:reconciled-state>n</split:reconciled-state>
      <split:value>-976509/100</split:value>
      <split:quantity>-976509/100</split:quantity>
      <split:account type="guid">6c736c815b25f0ab5b7b96571e15812b</split:account>
    </trn:split>
  </trn:splits>
</gnc:transaction>
<gnc:transaction version="2.0.0">
  <trn:id type="guid">1819adbc328c91036af266ace42d79b7</trn:id>
  <trn:currency>
    <cmdty:space>ISO4217</cmdty:space>
    <cmdty:id>CAD</cmdty:id>
  </trn:currency>
  <trn:date-posted>
    <ts:date>2007-12-12 05:59:00 -0500</ts:date>
  </trn:date-posted>
  <trn:date-entered>
    <ts:date>2010-02-14 23:50:26 -0500</ts:date>
  </trn:date-entered>
  <trn:description>Investment</trn:description>
  <trn:splits>
    <trn:split>
      <split:id type="guid">aeeeac5b036ac2ebb82dab8496761f71</split:id>
      <split:action>Buy</split:action>
      <split:reconciled-state>n</split:reconciled-state>
      <split:value>195300/100</split:value>
      <split:quantity>82612/1000</split:quantity>
      <split:account type="guid">b761bc5790b7d1018aef110c10921a45</split:account>
    </trn:split>
    <trn:split>
      <split:id type="guid">c8399cdbd9dd7261bf7db0ae8c72bd7b</split:id>
      <split:reconciled-state>n</split:reconciled-state>
      <split:value>-195300/100</split:value>
      <split:quantity>-195300/100</split:quantity>
      <split:account type="guid">53f6507e7492d6fb0600772d2da50cff</split:account>
    </trn:split>
  </trn:splits>
</gnc:transaction>
<gnc:transaction version="2.0.0">
  <trn:id type="guid">7a2196da44c93a6624d0de0e966bee82</trn:id>
  <trn:currency>
    <cmdty:space>ISO4217</cmdty:space>
    <cmdty:id>CAD</cmdty:id>
  </trn:currency>
  <trn:date-posted>
    <ts:date>2011-09-15 06:59:00 -0400</ts:date>
  </trn:date-posted>
  <trn:date-entered>
    <ts:date>2011-12-04 16:41:25 -0500</ts:date>
  </trn:date-entered>
  <trn:description>Sienna</trn:description>
  <trn:slots>
    <slot>
      <slot:key>notes</slot:key>
      <slot:value type="string"></slot:value>
    </slot>
  </trn:slots>
  <trn:splits>
    <trn:split>
      <split:id type="guid">68f256fd6f14624532a8d68879ad26ec</split:id>
      <split:reconciled-state>n</split:reconciled-state>
      <split:value>4500/100</split:value>
      <split:quantity>4500/100</split:quantity>
      <split:account type="guid">fed7736c5ecc376fabd1a96d6aac8218</split:account>
    </trn:split>
    <trn:split>
      <split:id type="guid">2808165b8c2653461123570bc2b49838</split:id>
      <split:reconciled-state>n</split:reconciled-state>
      <split:value>-4500/100</split:value>
      <split:quantity>-4500/100</split:quantity>
      <split:account type="guid">53f6507e7492d6fb0600772d2da50cff</split:account>
    </trn:split>
  </trn:splits>
</gnc:transaction>
<gnc:transaction version="2.0.0">
  <trn:id type="guid">3b41ae322857de47107fd0cd4c55151c</trn:id>
  <trn:currency>
    <cmdty:space>ISO4217</cmdty:space>
    <cmdty:id>CAD</cmdty:id>
  </trn:currency>
  <trn:date-posted>
    <ts:date>2011-09-18 06:59:00 -0400</ts:date>
  </trn:date-posted>
  <trn:date-entered>
    <ts:date>2011-12-04 16:42:10 -0500</ts:date>
  </trn:date-entered>
  <trn:description>Sienna</trn:description>
  <trn:slots>
    <slot>
      <slot:key>notes</slot:key>
      <slot:value type="string">OFX ext. info: |Trans type:Generic debit</slot:value>
    </slot>
  </trn:slots>
  <trn:splits>
    <trn:split>
      <split:id type="guid">24ceec6389edd89fa2628619dd677748</split:id>
      <split:reconciled-state>n</split:reconciled-state>
      <split:value>3751/100</split:value>
      <split:quantity>3751/100</split:quantity>
      <split:account type="guid">fed7736c5ecc376fabd1a96d6aac8218</split:account>
    </trn:split>
    <trn:split>
      <split:id type="guid">1c5cbff316fd6d56b8cc487352293b3f</split:id>
      <split:reconciled-state>n</split:reconciled-state>
      <split:value>-3751/100</split:value>
      <split:quantity>-3751/100</split:quantity>
      <split:account type="guid">53f6507e7492d6fb0600772d2da50cff</split:account>
    </trn:split>
  </trn:splits>
</gnc:transaction>
</gnc:book>
</gnc-v2>

<!-- Local variables: -->
<!-- mode: xml        -->
<!-- End:             -->
`)
