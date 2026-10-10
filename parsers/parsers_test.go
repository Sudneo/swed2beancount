package parsers

import (
	"testing"
	"time"

	"github.com/sudneo/swed2beancount/models"
)

func TestParseSwedbank(t *testing.T) {
	records := [][]string{
		{"Client account", "Row type", "Date", "Beneficiary/Payer", "Details", "Amount", "Currency", "Debit/Credit"},
		{"EE01", "10", "01.01.2021", "", "Opening balance", "100", "EUR", "K"},
		{"EE01", "20", "08.01.2021", "RIMI", "Order 1", "75,78", "EUR", "D"},
		{"EE01", "20", "09.01.2021", "EMPLOYER", "Salary", "300", "EUR", "K"},
		{"EE01", "82", "31.01.2021", "", "Turnover", "75,78", "EUR", "D"},
		{"EE01", "900", "31.01.2021", "", "Accrued interest", "0,01", "EUR", "K"},
		{"EE01", "86", "31.01.2021", "", "closing balance", "12,34", "EUR", "D"},
	}
	transactions, balances, err := parseSwedbank(records)
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 {
		t.Fatalf("Got %d transactions, expected 2: %+v", len(transactions), transactions)
	}
	if tr := transactions[0]; tr.ClientAccount != "EE01" || tr.Type != "debit" || tr.Amount != 7578 || tr.Beneficiary != "RIMI" {
		t.Errorf("Debit parsed incorrectly: %+v", tr)
	}
	if tr := transactions[1]; tr.Type != "credit" || tr.Amount != 30000 {
		t.Errorf("Credit parsed incorrectly: %+v", tr)
	}
	want := []models.Balance{
		{ClientAccount: "EE01", Date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), Amount: 10000, Currency: "EUR"},
		// The closing balance is asserted on the following day, negative if debit
		{ClientAccount: "EE01", Date: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC), Amount: -1234, Currency: "EUR"},
	}
	if len(balances) != len(want) {
		t.Fatalf("Got %d balances, expected %d: %+v", len(balances), len(want), balances)
	}
	for i := range want {
		if balances[i] != want[i] {
			t.Errorf("Balance %d: got %+v, expected %+v", i, balances[i], want[i])
		}
	}
}

func TestParseSwedbankInvalidRow(t *testing.T) {
	header := []string{"Client account", "Row type", "Date", "Beneficiary/Payer", "Details", "Amount", "Currency", "Debit/Credit"}
	for _, row := range [][]string{
		{"EE01", "20", "2021-01-08", "", "", "1", "EUR", "D"},
		{"EE01", "20", "08.01.2021", "", "", "abc", "EUR", "D"},
		{"EE01", "20"},
	} {
		if _, _, err := parseSwedbank([][]string{header, row}); err == nil {
			t.Errorf("Expected an error for row %v", row)
		}
	}
}

func TestParseStatementsNoMatch(t *testing.T) {
	if _, _, err := ParseStatements([]string{"../test/missing_*.csv"}); err == nil {
		t.Errorf("Expected an error when no statement matches")
	}
}
