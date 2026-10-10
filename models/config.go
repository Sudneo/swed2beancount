package models

import (
	"fmt"
	"os"
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
	CSVFile         string `yaml:"csv_file"`
	BeancountLedger string `yaml:"ledger"`
	StartDate       Date   `yaml:"from_date,omitempty"`
	EndDate         Date   `yaml:"until_date,omitempty"`
	DefaultAccount  string `yaml:"default_account,omitempty"`
}

func parseConfig(yamlData []byte) (Config, error) {
	config := Config{
		MappingFile:     "mappings.yaml",
		CSVFile:         "statement.csv",
		BeancountLedger: "ledger.beancount",
	}
	err := yaml.Unmarshal(yamlData, &config)
	return config, err
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
