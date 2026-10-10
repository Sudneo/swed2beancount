package models

import (
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	c := []byte(`ledger: "test-ledger.beancount"
csv_file: "swedbank_statement.csv"
mappings_file: "mappings.yaml"
output_file: "shared.beancount"
from_date: "2021-12-01"
until_date: "2021-12-30"`)
	config, err := parseConfig(c)
	if err != nil {
		t.Errorf("Parsing the configuration YAML lead to error")
	}
	if config.BeancountLedger != "test-ledger.beancount" {
		t.Errorf("Ledger parsed incorrectly. Got %s, expected %s", config.BeancountLedger, "test-ledger.beancount")
	}
	if config.MappingFile != "mappings.yaml" {
		t.Errorf("Mapping parsed incorrectly. Got %s, expected %s", config.MappingFile, "mappings.yaml")
	}
	if config.OutputFile != "shared.beancount" {
		t.Errorf("Outputfile parsed incorrectly. Got %s, expected %s", config.OutputFile, "shared.beancount")
	}
	csvExpected := "swedbank_statement.csv"
	if config.CSVFile != csvExpected {
		t.Errorf("CSVFile parsed incorrectly. Got %s, expected %s", config.CSVFile, csvExpected)
	}
	dateStartExpected, _ := time.Parse("2006-01-02", "2021-12-01")
	dateEndExpected, _ := time.Parse("2006-01-02", "2021-12-30")
	if !config.StartDate.Equal(dateStartExpected) || !config.EndDate.Equal(dateEndExpected) {
		t.Errorf("Date parsing is incorrect. Got %v - %v, expected %v - %v", config.StartDate, config.EndDate, dateStartExpected, dateEndExpected)
	}

}

func TestParseConfigWithoutDates(t *testing.T) {
	config, err := parseConfig([]byte(`ledger: "test-ledger.beancount"`))
	if err != nil {
		t.Fatalf("Parsing the configuration YAML lead to error: %v", err)
	}
	if !config.StartDate.IsZero() || !config.EndDate.IsZero() {
		t.Errorf("Dates should be zero when not specified. Got %v - %v", config.StartDate, config.EndDate)
	}
}

func TestParseConfigInvalidDate(t *testing.T) {
	for _, c := range []string{`from_date: "2021-13-01"`, `until_date: "01/12/2021"`} {
		if _, err := parseConfig([]byte(c)); err == nil {
			t.Errorf("Expected an error for invalid date in %q", c)
		}
	}
}

func TestReadConfig(t *testing.T) {
	filename := "../test/config.yaml"
	_, err := ReadConfig(filename)
	if err != nil {
		t.Errorf("Reading the configuration YAML lead to error")
	}
}
