Converts CSV files to coin transactions.

`csv2coin` is looking for the following values

* account - target account ID (corresponding accounts should be tagged with the `csv_acctid` directive)
* description - transaction description
* date - date of the transaction
* amount - the cost of the transaction
* symbol - (optional) symbol of the commodity that was traded
* quantity - (optional) quantity of the commodity that was traded
* note - (optional) note associated with the transaction

Conversion of individual CSV records (lines) to these values is driven by import rules. Import rules are usually described in `$COINDB/csv.rules` file. Alternatively, if the values can be directly lifted from the full contents of specific fields of the CSV records, this mapping can be provided directly on the commandline as a list of field indexs through the `-fields` option. The order of indexes follows the order of values in the list above, e.g. `-fields=3,0,2,6,1,7` for an import that won't have notes.

## csv.rules

The file consists of two sections separated with a single line of 3 dashes (`---\n`). Either section is optional.

The first section describes known CSV sources and the value mappings for each. Different institutions structure their CSV exports differently, so each is likely to require a different source mapping, possibly several (e.g. if the export doesn't include the account ID, a separate source for each account will be required). Each source is given a name and the source to use for given import is selected via the `-source` option.

The second section provides rules for picking the target accounts based on the transaction descriptions. It works exactly the same as described in [`ofx.rules`](https://github.com/mkobetic/coin/blob/master/cmd/ofx2coin/README.md#ofx.rules)

### source mapping

Each source description starts with a line naming the source and specifying the number of lines to skip before to get to the actual transaction records. The rest of the description consists of lines starting with whitespace and each line describing a rule for extracting a value from a record field or a rule for deriving a value it from the other values. Each line starts with a value name. There should be a line for each of the required values described above. 

#### extraction rules

An extraction rule starts with a field index indicating which field of the CSV record should be used. If that is the whole rule, the entire contets of the field is used as the value. The index can be followed by two expressions.

The first expression is a template enclosed in doublequotes composing the value from the results of the second expression. The second expression follows and continues until the end of the line and contains a regular expression with capturing groups. The first expression can reference the matched subgroups via their index.
The `amount` rule in the following example will match the third field of the record (indexes start from 0) with the expression `VALUE = (\d+)` and use subgroup 1 as the value ($0 is the match of the full expression).

```
source 2
  account 1
  description 3
  amount 2 "$1" VALUE = (\d+)
```

#### derivation rules

Derivation rule composes the value out of other values defined in the source. The rule has a single expression enclosed in doublequotes that is a template of the value referencing other values in the source by name. The example uses constant `123` as the account ID and composes the `note` value out of two separate record fields 1 and 0.

```
source 1
  account "123"
  description 3
  note1 1
  note2 0
  note "${note2} => ${note1}"
```