This package reads GnuCash XML database file (v2) and converts it to equivalent coin structures.
`cmd/gc2coin` shows how it's meant to be used.

## Implementation Notes

* The `encoding/xml` package is unable to unmarshal properly namespaced tags https://github.com/golang/go/issues/9519. However it is able to ignore the namespace prefixes if they are removed from the struct tags. Consequently we cannot marshal back into proper gnucash XML.