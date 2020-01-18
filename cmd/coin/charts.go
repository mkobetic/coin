package main

// Code generated from files in charts subdirectory. DO NOT EDIT.
//
//go:generate go run charts/embed.go

var charts = map[string][]byte{}

func init() {
	for file, encoded := range map[string]string{
    	"totals.html": "H4sIAAAAAAAA/0SQzW7rIBCF9zzF3NnH5CqKFFXApj/bdpEuuqSeKWA5tgUTt3n7ChOpK8SZ88GnMf+eXh/PH2/PEOUyOmXqAaOfgkWesAbsySkAc2Hx0EefC4vF9/PL7oSgt5EkGdnRoRsKjOzzlKZgdEvrvPQ5LQIl9xajyFIetKbDULo5B02Hbj12Q8G/x6/ytTuhM7qBThndNMznTDenlFkgJiKeLEq+MkIii+TFV2O91EZZA3wnkmjx/36/R4icQhSLx3qp/fqdbEBZw4Y0TbktbFH4R/TgV9/SOzITN+IupoxuSka3Bf4GAAD//zz+nINRAQAA",
    	"totals.js": "H4sIAAAAAAAA/2yQwUoDMRCG73mKn5XCDJa10puw+AQW8SbiYdhMaSCbls1sPMi+u6SuLULnFDI/3/8xRUZ4MUEHv237XF5lzPp2/Mrkt23WqL1Rc7qrmYbbgw2ReL2fUm/hmMjztwOAUW0aE3w7yIku27IOy75O2IMCug4bxvW3TpWwXwMLg54VqFm9P6xeGqbC/8JLlVVSmmLEMwqeYJfQDI1Zb1QkdLgvt2Ah72RHiRdWurLOr5ndzK4iDipex6oqJh+bT/d3OzFpcwy90iM79xMAAP//O+pDsFYBAAA=",
	} {
		charts[file] = decode(file, encoded)
	}
}
