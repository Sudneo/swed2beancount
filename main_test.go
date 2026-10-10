package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sudneo/swed2beancount/models"
)

func date(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestFilterTransactionsByDate(t *testing.T) {
	transactions := []models.Transaction{
		{Date: date("2021-09-30")},
		{Date: date("2021-10-01")},
		{Date: date("2021-12-31")},
		{Date: date("2022-01-01")},
	}
	cases := []struct {
		name       string
		start, end time.Time
		want       int
	}{
		{"both bounds, inclusive", date("2021-10-01"), date("2021-12-31"), 2},
		{"start only", date("2021-10-01"), time.Time{}, 3},
		{"end only", time.Time{}, date("2021-12-31"), 3},
		{"no bounds", time.Time{}, time.Time{}, 4},
	}
	for _, c := range cases {
		if got := filterTransactionsByDate(c.start, c.end, transactions); len(got) != c.want {
			t.Errorf("%s: got %d transactions, expected %d", c.name, len(got), c.want)
		}
	}
}

var accounts = map[string]string{
	"EE01": "Assets:Bank:Checking",
	"EE02": "Assets:Bank:Shared",
}

func TestProcessTransactions(t *testing.T) {
	transactions := []models.Transaction{
		{ClientAccount: "EE01", Type: "debit", Beneficiary: "RIMI", Description: "Order 1"},
		{ClientAccount: "EE02", Type: "credit", Beneficiary: "EMPLOYER", Description: "salary"},
	}
	m := models.Mappings{{Contains: "rimi", Account: "Expenses:Food", DescriptionOverride: "Groceries"}}
	got, err := processTransactions(transactions, accounts, m)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].DebitAccount != "Assets:Bank:Checking" || got[0].CreditAccount != "Expenses:Food" || got[0].Description != "Groceries" {
		t.Errorf("Debit processed incorrectly: %+v", got[0])
	}
	if got[1].DebitAccount != uncategorizedAccount || got[1].CreditAccount != "Assets:Bank:Shared" || got[1].Description != "salary" {
		t.Errorf("Credit processed incorrectly: %+v", got[1])
	}
}

func TestProcessTransactionsUnknownAccount(t *testing.T) {
	transactions := []models.Transaction{{ClientAccount: "EE99", Type: "debit"}}
	if _, err := processTransactions(transactions, accounts, nil); err == nil {
		t.Errorf("Expected an error for an unconfigured account")
	}
}

func TestProcessBalances(t *testing.T) {
	balances := []models.Balance{
		{ClientAccount: "EE02", Date: date("2021-01-01")},
		{ClientAccount: "EE01", Date: date("2021-01-01")},
		{ClientAccount: "EE01", Date: date("2021-03-01")},
		{ClientAccount: "EE01", Date: date("2021-03-02")},
	}
	got, err := processBalances(date("2021-01-01"), date("2021-02-28"), balances, accounts)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("Got %d balances, expected 3: %+v", len(got), got)
	}
	if got[0].Account != "Assets:Bank:Checking" || got[1].Account != "Assets:Bank:Shared" || !got[2].Date.Equal(date("2021-03-01")) {
		t.Errorf("Balances processed incorrectly: %+v", got)
	}
	if _, err := processBalances(time.Time{}, time.Time{}, []models.Balance{{ClientAccount: "EE99"}}, accounts); err == nil {
		t.Errorf("Expected an error for an unconfigured account")
	}
}

func TestRunRequiresAccounts(t *testing.T) {
	config := models.Config{
		BeancountLedger: "test/ledger.beancount",
		Statements:      []string{"test/swedbank_statement.csv"},
	}
	if err := run(config); err == nil {
		t.Errorf("Expected an error without accounts")
	}
}

func TestRun(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.beancount")
	config := models.Config{
		MappingFile:     "test/mappings.yaml",
		Statements:      []string{"test/swedbank_*.csv"},
		BeancountLedger: "test/ledger.beancount",
		OutputFile:      out,
		Accounts: map[string]string{
			"EE000000000000000000": "Assets:Bank:Personal:Checking",
			"EE000000000000000001": "Assets:Bank:Shared:Checking",
		},
	}
	if err := run(config); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	want := `
2021-01-08 * "Grocery - Rimi"
	Expenses:Food:Groceries 75.78 EUR
	Assets:Bank:Personal:Checking -75.78 EUR

2021-02-05 * "MAXIMA Tallinn "
	UNCATEGORIZED_ACCOUNT 43.58 EUR
	Assets:Bank:Personal:Checking -43.58 EUR

2021-02-25 * "Monthly Salary"
	Assets:Bank:Personal:Checking 300.00 EUR
	Income:Salary -300.00 EUR

2021-02-27 * "Kaardi kuutasu 02.2021"
	UNCATEGORIZED_ACCOUNT 0.40 EUR
	Assets:Bank:Personal:Checking -0.40 EUR

2021-02-28 * "Monthly contribution"
	Assets:Bank:Shared:Checking 50.00 EUR
	UNCATEGORIZED_ACCOUNT -50.00 EUR

2021-02-28 * "Monthly contribution"
	Assets:Bank:Shared:Checking 50.00 EUR
	Assets:Bank:Personal:Checking -50.00 EUR

2021-01-01 balance Assets:Bank:Personal:Checking 100.00 EUR
2021-01-01 balance Assets:Bank:Shared:Checking 300.00 EUR
2021-03-01 balance Assets:Bank:Personal:Checking 230.24 EUR
2021-03-01 balance Assets:Bank:Shared:Checking 350.00 EUR
`
	if string(got) != want {
		t.Errorf("Unexpected output:\n%s", got)
	}
}
