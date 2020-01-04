Converts CSV files to coin transactions.

Requires a field for each of the following values

* account - target account ID
* description - transaction description
* date - date of the transaction
* amount - the cost of the transaction
* symbol - symbol of the commodity that was traded
* quantity - quantity of the commodity that was traded
* note - optional note associated with the transaction

The fields are specified as field indexes in the CSV records. Indexes can be specified directly on the command line through the `-fields` option. Alternatively they can be attached to a named "source" in the `csv.rules` file and selected through the `-source` option.