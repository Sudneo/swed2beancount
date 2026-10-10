package models

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	yaml "gopkg.in/yaml.v3"
)

// Date is a calendar date read from YAML in YYYY-MM-DD format.
type Date struct{ time.Time }

func (d *Date) UnmarshalYAML(n *yaml.Node) error {
	t, err := time.Parse("2006-01-02", n.Value)
	if err != nil {
		return fmt.Errorf("invalid date %q (want YYYY-MM-DD): %w", n.Value, err)
	}
	d.Time = t
	return nil
}

type Config struct {
	MappingFile     string `yaml:"mappings_file"`
	OutputFile      string `yaml:"output_file"`
	BeancountLedger string `yaml:"ledger"`
	StartDate       Date   `yaml:"from_date,omitempty"`
	EndDate         Date   `yaml:"until_date,omitempty"`
	// Statements are the CSV files to import, as paths or glob patterns.
	Statements []string `yaml:"statements"`
	// Accounts maps the IBAN of each account to its beancount account.
	Accounts map[string]string `yaml:"accounts"`
}

func parseConfig(yamlData []byte) (Config, error) {
	config := Config{
		MappingFile:     "mappings.yaml",
		BeancountLedger: "ledger.beancount",
	}
	// Reject unknown keys, so that old options (e.g., "csv_file" or
	// "default_account") are reported.
	decoder := yaml.NewDecoder(bytes.NewReader(yamlData))
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil && !errors.Is(err, io.EOF) {
		return config, err
	}
	return config, nil
}

func ReadConfig(configFile string) (Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}
	config, err := parseConfig(data)
	if err != nil {
		return Config{}, fmt.Errorf("parsing config %s: %w", configFile, err)
	}
	return config, nil
}

// Validate checks that statements and accounts are configured and that the
// accounts are opened in the ledger.
func (c Config) Validate(ledgerAccounts []string) error {
	if len(c.Statements) == 0 {
		return errors.New("no statements configured")
	}
	if len(c.Accounts) == 0 {
		return errors.New("no accounts configured")
	}
	for iban, account := range c.Accounts {
		if !slices.Contains(ledgerAccounts, account) {
			return fmt.Errorf("account %q for %s is not opened in the ledger", account, iban)
		}
	}
	return nil
}
