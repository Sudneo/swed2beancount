package models

import (
	"errors"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Config struct {
	MappingFile       string `yaml:"mappings_file"`
	OutputFile        string `yaml:"output_file"`
	CSVFile           string `yaml:"csv_file"`
	CSVType           string `yaml:"csv_type"`
	BeancountLedger   string `yaml:"ledger"`
	StartDate         string `yaml:"from_date,omitempty"`
	EndDate           string `yaml:"until_date,omitempty"`
	InternalStartDate time.Time
	InternalEndDate   time.Time
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
	if config.StartDate != "" {
		layout := "2006-01-02"
		d, err := time.Parse(layout, config.StartDate)
		if err != nil {
			log.WithFields(log.Fields{
				"Field": "StartDate",
				"Value": config.StartDate,
			}).Error("Invalid date specified")
			return config, errors.New("Failed to validate config")
		}
		config.InternalStartDate = d
	}
	if config.EndDate != "" {
		layout := "2006-01-02"
		d, err := time.Parse(layout, config.EndDate)
		if err != nil {
			log.WithFields(log.Fields{
				"Field": "EndDate",
				"Value": config.EndDate,
			}).Error("Invalid date specified")
			return config, errors.New("Failed to validate config")
		}
		config.InternalEndDate = d
	}
	return config, nil

}

func ReadConfig(configFile string) (Config, error) {
	var config Config
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Could not read config file")
		return config, errors.New("Failed to validate config file")
	}
	config, err = parseConfig(yamlFile)
	return config, err
}
