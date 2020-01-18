package main

// Code generated from files in charts subdirectory. DO NOT EDIT.
//
//go:generate go run charts/embed.go

var charts = map[string][]byte{}

func init() {
	for file, encoded := range map[string]string{
    	"totals.html": "H4sIAAAAAAAA/0SQzW7rIBCF9zzF3NnH5CqKFFXApj/bdpEuuqSeKWA5tgUTt3n7ChOpK8SZ88GnMf+eXh/PH2/PEOUyOmXqAaOfgkWesAbsySkAc2Hx0EefC4vF9/PL7oSgt5EkGdnRoRsKjOzzlKZgdEvrvPQ5LQIl9xajyFIetKbDULo5B02Hbj12Q8G/x6/ytTuhM7qBThndNMznTDenlFkgJiKeLEq+MkIii+TFV2O91EZZA3wnkmjx/36/R4icQhSLx3qp/fqdbEBZw4Y0TbktbFH4R/TgV9/SOzITN+IupoxuSka3Bf4GAAD//zz+nINRAQAA",
    	"totals.js": "H4sIAAAAAAAA/5xUwY7aMBC95ytGXq1ki+AN2lsRqnbbY7utql4qxMFNTLDWcZA9pMmu+PfKdsImFKqqPgD2vPfmeYZxIywUAgWsoLjnuWu+CuskLe65k1rmSMn+xscJ4yhbpCz1OHHA+nu3lyzxfNeUkT5QXFPe5DthkUTATqpyh7CCmWtKLhAtJfGsB/xSBe6m8XDUh9tePhdaPgpTUMatMKWk6yyN3E0EdiPgJ2WksG/QmDCFrMe+jLBfbKGM0OHa+U5W8oNAWda2W2RsGS08tMpFhmiVe6wR64q2fdpp8JtPRDuWhKAWnbR91KHInynjz7Jz1JeV57U+VMZxp1Uu6YKxcMySpOVFXQllaOSvsw2vxJ5uDyZHVRtaMHgFK/FgDRQ8aH0UKJdwZGyZdAPbV6i455VoB6H4xbU0Je5gDotNCpdV14sNHNmGceO9seXoPrDyXe8b/qA1JTycE5YAAAQ/fcL+RBqUvhtiv5emoKQcoLHduRbOkRTIRMZhpyUlW6U1GbtMQY2MvlDF/L2XSRLYY1tW5jgxdfmqnn7Z50jgZLYlV0rW0lEr2En0ROyuETtarLPNBUY/Jn+nwTz8WlwSiHOUQst/ClOEHfWEhS8XhC6et+TPpvh/NfiP+bw9B6EVxm1rW3lg2GiBkmYpgdkw+TMg7I2XC61pGKj/9ND9owfvID4tMyBpdm6hGywkd3fj+r76vV+n8ZpMXpOqN4hfagtUwWoFGYNJwC8/MhinH1Ul4/NKbn/c3X4mjDbsHN/nRK9nDlrDe2jgHeAYdwSpnbycy/hntLmiqtyTeKKG9aJmIjpsjsHTkf0OAAD//8Z7F9EbBgAA",
	} {
		charts[file] = decode(file, encoded)
	}
}
