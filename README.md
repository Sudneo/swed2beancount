# Beancount CSV Parser

This simple tool is just some util I wrote to help importing CSV files into beancount format. The tool only supports the Swedbank CSV statement export.

## Features

* Imports the statements of multiple accounts at once
* Mapping between CSV transaction and Beancount accounts
    * Text match (on remitter/beneficiary and/or on the details/description)
* Duplicate transactions (e.g., transfers between your own accounts, which appear in both statements) are imported once
* Balance assertions generated from the opening and closing balances of the statements
* Automatically generate beancount-like file 
* Validates data in configuration and mappings (e.g., invalid account name)
* Select date range for the transactions to consider (.e.g, month by month on a full-year export)

## How to Use

1. Write a configuration file:

```yaml
ledger: "ledger.beancount"
mappings_file: "mappings.yaml"
output_file: "import.beancount"
statements:
  - "downloads/*.csv"
from_date: "2021-10-01"
until_date: "2021-12-31"
accounts:
  EE000000000000000001: "Assets:Bank:Personal:Checking"
  EE000000000000000002: "Assets:Bank:Shared:Checking"
  EE000000000000000003: "Liabilities:Bank:CreditCard"
  EE000000000000000004: "Assets:Bank:Personal:Investing"
```

This will use the specified ledger to source accounts (used for validation) but it **will not** modify it.
The output will be placed in the file specified in `output_file`.

`statements` lists the CSV exports to import, as paths or glob patterns. Each file can contain one or more accounts.

`accounts` maps the IBAN of each account (the `Client account` column of the CSV) to the beancount account that is the implicit 'leg' of its transactions (+ for credits, - for debits). Every account found in the statements must be listed here.

`from_date` and `until_date` are inclusive and each can be omitted to leave that side of the range open. `ledger` defaults to `ledger.beancount` and `mappings_file` to `mappings.yaml`.

2. Balance assertions

The opening and closing balances of each statement are written as `balance` directives after the transactions. Since beancount checks balances at the beginning of the day, the closing balance of a day is asserted on the following one. Only the balances on the edges of the `from_date`/`until_date` range are kept: to get both, export statements covering the same range.

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
- contains: "JOINT ACCOUNT"
  field: "beneficiary"
  account: "Assets:Bank:Shared:Checking"
```

Each rule matches transactions that contain some string in the remitter/beneficiary or in the description. The matching is case-insensitive and is done from top to bottom, which means that if multiple rules match, the first match will be the used one.

Each rule has the following fields:

* `contains`: The string to search
* `field` (optional): `beneficiary` or `details` to restrict the matching to one of the two fields only. If not specified, both fields will be evaluated.

The `account` field always needs to be specified and needs to match a beancount account present in the ledger specified in the config file.

A rule can map to one of your own accounts (e.g., the shared account) to categorize transfers between them. Such a transfer appears in both statements: when both sides are mapped to the other account, they result in two transactions with the same date, amount and accounts, and only the first one is kept.

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
./csv2beancount [-config config.yaml] [-v]
```
