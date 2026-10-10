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

func TestProcessTransactions(t *testing.T) {
	transactions := []models.Transaction{
		{Type: "debit", Beneficiary: "RIMI", Description: "Order 1"},
		{Type: "credit", Beneficiary: "EMPLOYER", Description: "salary"},
	}
	m := models.Mappings{{Contains: "rimi", Account: "Expenses:Food", DescriptionOverride: "Groceries"}}
	got := processTransactions(transactions, "Assets:Bank", m)
	if got[0].DebitAccount != "Assets:Bank" || got[0].CreditAccount != "Expenses:Food" || got[0].Description != "Groceries" {
		t.Errorf("Debit processed incorrectly: %+v", got[0])
	}
	if got[1].DebitAccount != uncategorizedAccount || got[1].CreditAccount != "Assets:Bank" || got[1].Description != "salary" {
		t.Errorf("Credit processed incorrectly: %+v", got[1])
	}
}

func TestRunRequiresDefaultAccount(t *testing.T) {
	if err := run(models.Config{}); err == nil {
		t.Errorf("Expected an error without a default account")
	}
}

func TestRun(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.beancount")
	config := models.Config{
		MappingFile:     "test/mappings.yaml",
		CSVFile:         "test/swedbank_statement.csv",
		BeancountLedger: "test/ledger.beancount",
		OutputFile:      out,
		DefaultAccount:  "Assets:Bank:Personal:Checking",
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

2021-02-27 * "Kaardi kuutasu 02.2021"
	UNCATEGORIZED_ACCOUNT 0.4 EUR
	Assets:Bank:Personal:Checking -0.4 EUR 

`
	if string(got) != want {
		t.Errorf("Unexpected output:\n%s", got)
	}
}
