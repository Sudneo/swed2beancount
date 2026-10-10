package models

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestReadAccounts(t *testing.T) {
	ledger := filepath.Join(t.TempDir(), "ledger.beancount")
	err := os.WriteFile(ledger, []byte(`; 2020-01-01 open Expenses:Commented
2020-01-01 open Assets:Bank:N26 EUR
2020-01-01 open Equity:Opening-Balances
2020-01-01  open  Expenses:Café
2020-01-01 * "Groceries"
    Expenses:Food 10 EUR
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	accounts, err := ReadAccounts(ledger)
	if err != nil {
		t.Fatalf("Reading the ledger lead to error: %v", err)
	}
	want := []string{"Assets:Bank:N26", "Equity:Opening-Balances", "Expenses:Café"}
	if !slices.Equal(accounts, want) {
		t.Errorf("Got %v, expected %v", accounts, want)
	}
}

func TestReadAccountsMissingFile(t *testing.T) {
	if _, err := ReadAccounts("does-not-exist.beancount"); err == nil {
		t.Errorf("Expected an error for a missing ledger")
	}
}
