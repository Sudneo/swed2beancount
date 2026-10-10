package models

import (
	"strings"
	"testing"
)

func TestReadMapping(t *testing.T) {
	m, err := ReadMapping("../test/mappings.yaml")
	if err != nil {
		t.Fatalf("Reading the mappings YAML lead to error: %v", err)
	}
	if len(m) == 0 {
		t.Errorf("No mappings parsed")
	}
}

func TestParseMappingsRejectsRemovedOptions(t *testing.T) {
	for _, c := range []string{
		"- type: text\n  contains: \"RIMI\"\n  account: \"Expenses:Food\"",
		"- exact_amount: 457.89\n  account: \"Expenses:Rent\"",
		"- date_begin: \"2021-06-01\"\n  date_end: \"2021-07-01\"\n  account: \"Expenses:Trip\"",
	} {
		if _, err := parseMappings([]byte(c)); err == nil {
			t.Errorf("Expected an error for unknown key in %q", strings.SplitN(c, "\n", 2)[0])
		}
	}
}

func TestValidate(t *testing.T) {
	accounts := []string{"Expenses:Food"}
	cases := []struct {
		name    string
		mapping Mapping
		wantErr bool
	}{
		{"valid", Mapping{Contains: "RIMI", Account: "Expenses:Food"}, false},
		{"valid with field", Mapping{Contains: "RIMI", Field: "details", Account: "Expenses:Food"}, false},
		{"empty contains", Mapping{Account: "Expenses:Food"}, true},
		{"unknown field", Mapping{Contains: "RIMI", Field: "amount", Account: "Expenses:Food"}, true},
		{"unknown account", Mapping{Contains: "RIMI", Account: "Expenses:Nope"}, true},
	}
	for _, c := range cases {
		m := Mappings{c.mapping}
		err := m.Validate(accounts)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: got error %v, wantErr %v", c.name, err, c.wantErr)
		}
	}
}
