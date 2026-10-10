package models

import "testing"

func TestCategorizeField(t *testing.T) {
	tr := Transaction{Beneficiary: "SOMEONE", Description: "paid at HACKTHEBOX"}
	cases := []struct {
		field string
		want  string
	}{
		{"beneficiary", ""},
		{"details", "Expenses:HTB"},
		{"", "Expenses:HTB"},
	}
	for _, c := range cases {
		m := Mappings{{Contains: "hackthebox", Field: c.field, Account: "Expenses:HTB"}}
		got, _ := tr.Categorize(m)
		if got != c.want {
			t.Errorf("field %q: got %q, expected %q", c.field, got, c.want)
		}
	}
}

func TestCategorizeFirstMatchWins(t *testing.T) {
	tr := Transaction{Beneficiary: "RIMI EESTI FOOD AS", Description: "Order 1"}
	m := Mappings{
		{Contains: "rimi", Field: "beneficiary", Account: "Expenses:Food", DescriptionOverride: "Grocery"},
		{Contains: "order", Account: "Expenses:Other"},
	}
	account, desc := tr.Categorize(m)
	if account != "Expenses:Food" || desc != "Grocery" {
		t.Errorf("Got %q/%q, expected %q/%q", account, desc, "Expenses:Food", "Grocery")
	}
}
