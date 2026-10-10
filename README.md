# Beancount CSV Parser

This simple tool is just some util I wrote to help importing CSV files into beancount format. The tool was initially specifically written for Swedbank CSV export, but I made it so that it can be used for different formats as well, with little code modification.

## Features

* Mapping between CSV transaction and Beancount accounts
    * Text match (on remitter/beneficiary and/or on the details/description)
* Automatically generate beancount-like file 
* Validates data in configuration and mappings (e.g., invalid account name)
* Select date range for the transactions to consider (.e.g, month by month on a full-year export)

## How to Use

1. Write a configuration file:

```yaml
ledger: "ledger.beancount"
csv_file: "statement.csv"
csv_type: "swedbank"
mappings_file: "mappings.yaml"
output_file: "shared.beancount"
from_date: "2021-10-01"
until_date: "2021-12-31"
```

This will use the specified ledger to source accounts (used for validation) but it **will not** modify it.
The output will be placed in the file specified in `output_file`.

Dates can be omitted. Most of the values have defaults:

```yaml
ledger: ledger.beancount
mappings_file: mappings.yaml
csv_file: statement.csv
csv_type: swedbank
```

2. Choose the default account to use for transactions. Since this is supposed to be a CSV parser, it is assumed that the export is for one specific account (e.g., Assets:Bank:Checking). This is the account that will be the implicit 'leg' of each transaction (+ for credits, - for debits).

3. Create a mapping file with rules such as:

```yaml
- contains: "HACKTHEBOX"
  field: "beneficiary"
  account: "Expenses:HTB"
  desc_override: "Hackthebox Subscription"
- contains: "RIMI EESTI FOOD AS"
  field: "details"
  account: "Expenses:Food:Groceries"
  desc_override: "Grocery - Rimi"
```

Each rule matches transactions that contain some string in the remitter/beneficiary or in the description. The matching is case-insensitive and is done from top to bottom, which means that if multiple rules match, the first match will be the used one.

Each rule has the following fields:

* `contains`: The string to search
* `field` (optional): `beneficiary` or `details` to restrict the matching to one of the two fields only. If not specified, both fields will be evaluated.

The `account` field always needs to be specified and needs to match a beancount account present in the ledger specified in the config file.

Additionally, the `desc_override` can be set to override the description for the transaction that matches that particular rule. If not specified, the details/description available in the CSV will be used.

4. Run the tool

To run the tool it is possible either to compile it or to download a prebuilt binary.

```bash
git clone https://github.com/sudneo/csv2beancount
cd csv2beancount
go build
```

Then it is possible to run the tool as follows:

```bash
./csv2beancount [-config config.yaml] -a "Assets:Bank:Checking"
```

## Additional Parsers

A new parser can be easily added by adding one more case to the `parsers.go` file.
It is expected that the following data will be availabe:

* Date
* Beneficiary
* Details
* Amount
* Currency
* Type (debit/credit)

The new parsing function will need to populate a list of `models.Transaction` with the data from the CSV. The logic can be implemented completely independent.
For example, if the CSV export contains amounts with positive and negative signs, this can be used to generate transaction with the positive amount and manually set the 'type' to 'credit'/'debit' depending on the sign.
