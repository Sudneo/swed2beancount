package models

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
	"os"
	"time"
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
	CSVType         string `yaml:"csv_type"`
	BeancountLedger string `yaml:"ledger"`
	StartDate       Date   `yaml:"from_date,omitempty"`
	EndDate         Date   `yaml:"until_date,omitempty"`
	DefaultAccount  string `yaml:"default_account,omitempty"`
}

func parseConfig(yamlData []byte) (Config, error) {
	var config Config
	err := yaml.Unmarshal(yamlData, &config)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Could not parse config file.")
		return config, errors.New("Failed to validate config file")
	}
	// Set defaults
	if config.MappingFile == "" {
		config.MappingFile = "mappings.yaml"
	}
	if config.CSVFile == "" {
		config.CSVFile = "statement.csv"
	}
	if config.CSVType == "" {
		config.CSVType = "swedbank"
	}
	if config.BeancountLedger == "" {
		config.BeancountLedger = "ledger.beancount"
	}
	return config, nil

}

func ReadConfig(configFile string) (Config, error) {
	var config Config
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Could not read config file")
		return config, errors.New("Failed to read config file")
	}
	config, err = parseConfig(yamlFile)
	return config, err
}
